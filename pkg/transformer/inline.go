// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transformer

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"context"
	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

// FileReader provides a generic file-reading interface for the Inliner to use.
type FileReader interface {
	ReadFile(ctx context.Context, file *bpb.File) ([]byte, error)
}

// Inliner inlines bundle files by reading them from the local or a remote
// filesystem.
type Inliner struct {
	// Local reader reads from the local filesystem.
	LocalReader FileReader
}

// NewInliner creates a new inliner. If the bundle is stored on disk, the cwd
// should be the absolute path to the directory containing the bundle file on disk.
func NewInliner(cwd string) *Inliner {
	return &Inliner{
		LocalReader: &localReader{cwd},
		// TODO: Add more readers for remote filesystems here.
	}
}

// localReader reads from the local filesystem.
type localReader struct {
	WorkingDir string
}

func (l *localReader) ReadFile(ctx context.Context, file *bpb.File) ([]byte, error) {
	path := file.GetPath()
	if path == "" {
		return nil, fmt.Errorf("file %v was specified but no file path was provided", file)
	}
	return ioutil.ReadFile(filepath.Join(l.WorkingDir, path))
}

// Inline converts dereferences File-references in a bundle and turns them into
// inline-references. Thus, the returned bundle is a copy with the
// file-references removed. This doesn't apply to binary images, which are left
// as-is.
func (n *Inliner) Inline(ctx context.Context, b *bpb.ClusterBundle) (*bpb.ClusterBundle, error) {
	// TODO: Rewrite this so it can be inlined concurrently. This is less
	// important for inlining local files and more important for inlining
	// external files.
	b = converter.CloneBundle(b)
	spec := b.GetSpec()
	if spec == nil {
		// Everything is trivially inlined!
		return b, nil
	}
	// Process all node-bootstrap files.
	for _, v := range spec.GetImageConfigs() {
		k := v.GetName()
		if v.GetInitScriptFile() != nil {
			contents, err := n.ReadFile(ctx, v.GetInitScriptFile())
			if err != nil {
				return nil, fmt.Errorf("error processing init script for node bootstrap config %q: %v", k, err)
			}
			v.InitData = &bpb.ImageConfig_InitScript{string(contents)}
		}
	}

	// Process all the cluster applications files.
	for _, v := range spec.GetClusterApps() {
		k := v.GetName()
		if err := n.processClusterApp(ctx, k, v); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func (n *Inliner) processClusterApp(ctx context.Context, k string, b *bpb.ClusterApplication) error {
	for _, co := range b.GetClusterObjects() {
		kco := co.GetName()
		if co.GetFile() != nil {
			contents, err := n.ReadFile(ctx, co.GetFile())
			if err != nil {
				return fmt.Errorf("error reading cluster app object for app %q and object %q: %v", k, kco, err)
			}
			pb, err := converter.Struct.YAMLToProto(contents)
			if err != nil {
				return fmt.Errorf("error converting cluster app object for app %q and object %q: %v", k, kco, err)
			}
			co.KubeData = &bpb.ClusterObject_Inlined{converter.ToStruct(pb)}
		}

		if co.GetOverlayFile() != nil {
			contents, err := n.ReadFile(ctx, co.GetOverlayFile())
			if err != nil {
				return fmt.Errorf("error reading overlay file for app %q and object %q: %v", k, kco, err)
			}
			pb, err := converter.OverlayCollection.YAMLToProto(contents)
			if err != nil {
				return fmt.Errorf("error converting overlay collection for app %q and object %q: %v", k, kco, err)
			}
			co.OverlayData = &bpb.ClusterObject_OverlayCollection{converter.ToOverlayCollection(pb)}
		}
	}
	return nil
}

// ReadFile from either a local or remote location.
func (n *Inliner) ReadFile(ctx context.Context, file *bpb.File) ([]byte, error) {
	path := file.GetPath()
	if path == "" {
		return nil, fmt.Errorf("file %v was specified but no file path was provided", file)
	}

	switch {
	case strings.HasPrefix("gs://", path):
		return nil, fmt.Errorf("path-type (GCS) not supported yet; file was %q", path)
	case strings.HasPrefix("https://", path) || strings.HasPrefix("http://", path):
		return nil, fmt.Errorf("path-type (HTTP[S]) not supported yet; file was %q", path)
	default:
		// By default, assume that the user expects to read from the local filesystem.
		return n.LocalReader.ReadFile(ctx, file)
	}
}
