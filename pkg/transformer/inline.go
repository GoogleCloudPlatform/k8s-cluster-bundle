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
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/golang/protobuf/proto"
)

// Inliner inlines bundle files by reading them from the local or a remote
// filesystem.
type Inliner struct {
	// Local reader reads from the local filesystem.
	Reader files.FilePBReader
}

// InlineOptions are options to give to the inliner.
type InlineOptions struct {
	// TopLayerOnly inlines the components and node configuration files and stops
	// there.
	TopLayerOnly bool
}

// NewInliner creates a new inliner. If the bundle is stored on disk, the cwd
// should be the absolute path to the directory containing the bundle file on disk.
func NewInliner(cwd string) *Inliner {
	return &Inliner{
		Reader: &files.LocalFilePBReader{cwd, &files.LocalFileSystemReader{}},
		// TODO: Add more readers for remote filesystems here or add some way to
		// plug-in other reader types.
	}
}

// Inline converts dereferences File-references in a bundle and turns them into
// inline-references. Thus, the returned bundle is a copy with the
// file-references removed. This doesn't apply to binary images, which are left
// as-is.
func (n *Inliner) Inline(ctx context.Context, b *bpb.ClusterBundle, opt *InlineOptions) (*bpb.ClusterBundle, error) {
	// TODO: Rewrite this so it can be inlined concurrently. This is less
	// important for inlining local files and more important for inlining
	// external files.
	b = converter.CloneBundle(b)
	spec := b.GetSpec()
	if spec == nil {
		// Everything is trivially inlined!
		return b, nil
	}

	// First, process any cluster component files or node config files.
	if err := n.processComponentPackageFiles(ctx, b); err != nil {
		return nil, err
	}

	if opt.TopLayerOnly {
		return b, nil
	}

	// Process all the cluster object files.
	for _, v := range spec.GetComponents() {
		k := v.GetMetadata().GetName()
		if err := n.processClusterObjects(ctx, k, v); err != nil {
			return nil, err
		}
	}
	return b, nil
}

var multiDoc = regexp.MustCompile("---(\n|$)")

func (n *Inliner) readFileToProto(ctx context.Context, file *bpb.File, conv *converter.Converter) ([]proto.Message, error) {
	contents, err := n.readFilePB(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %v", file.GetUrl(), err)
	}
	ext := filepath.Ext(file.GetUrl())

	// toYaml converts a possibly multi-doc YAML into proto.
	var empty []proto.Message
	toYaml := func(cont []byte) ([]proto.Message, error) {
		var out []proto.Message
		splat := multiDoc.Split(string(cont), -1)
		for i, s := range splat {
			c, err := conv.YAMLToProto([]byte(s))
			if err != nil {
				return empty, fmt.Errorf("error in document (%d) for file %q: %v", i, file.GetUrl(), err)
			}
			out = append(out, c)
		}
		return out, nil
	}

	switch ext {
	case ".yaml":
		return toYaml(contents)
	case ".json":
		z, err := conv.JSONToProto(contents)
		if err != nil {
			return empty, err
		}
		return []proto.Message{z}, nil
	default:
		return toYaml(contents)
	}
}

func (n *Inliner) processComponentPackageFiles(ctx context.Context, b *bpb.ClusterBundle) error {
	if b.Spec == nil {
		return nil
	}
	for _, cf := range b.GetSpec().GetComponentFiles() {
		pbs, err := n.readFileToProto(ctx, cf, converter.ComponentPackage)
		if err != nil {
			return fmt.Errorf("error reading component: %v", err)
		}
		for _, pb := range pbs {
			comp := converter.ToComponentPackage(pb)
			compName := comp.GetMetadata().GetName()
			if compName == "" {
				return fmt.Errorf("no component name (metadata.name) found for component with url %q",
					cf.GetUrl())
			}

			// Transform the file-urls of the children to now be relative to where the component is
			for _, ocf := range comp.GetSpec().GetClusterObjectFiles() {
				compUrl := cf.GetUrl()
				objUrl := ocf.GetUrl()
				if strings.HasPrefix(objUrl, "file://") && strings.HasPrefix(compUrl, "file://") {
					ocf.Url = "file://" + filepath.Join(filepath.Dir(shortFileUrl(compUrl)), shortFileUrl(objUrl))
				}
			}
			b.GetSpec().Components = append(b.GetSpec().Components, comp)
		}
	}
	var emptyFiles []*bpb.File
	b.GetSpec().ComponentFiles = emptyFiles
	return nil
}

func (n *Inliner) processClusterObjects(ctx context.Context, compName string, b *bpb.ComponentPackage) error {
	if b.Spec == nil {
		return nil
	}
	for _, cf := range b.GetSpec().GetClusterObjectFiles() {
		pbs, err := n.readFileToProto(ctx, cf, converter.Struct)
		if err != nil {
			return err
		}
		for _, pb := range pbs {
			if err != nil {
				return fmt.Errorf("error reading component object for component %q: %v", compName, err)
			}
			st := converter.ToStruct(pb)
			if len(st.GetFields()) > 0 {
				// Ignore 0-length objects (empty documents).
				b.GetSpec().ClusterObjects = append(b.GetSpec().ClusterObjects, st)
			}
		}
	}
	var emptyFiles []*bpb.File
	b.GetSpec().ClusterObjectFiles = emptyFiles
	return nil
}

// readFilePB from either a local or remote location.
func (n *Inliner) readFilePB(ctx context.Context, file *bpb.File) ([]byte, error) {
	url := file.GetUrl()
	if url == "" {
		return nil, fmt.Errorf("file %v was specified but no file path/url was provided", file)
	}

	switch {
	case strings.HasPrefix("gs://", url):
		return nil, fmt.Errorf("url-type (GCS) not supported; file was %q", url)
	case strings.HasPrefix("https://", url) || strings.HasPrefix("http://", url):
		return nil, fmt.Errorf("url-type (HTTP[S]) not supported; file was %q", url)
	case strings.HasPrefix("file://", url):
		return n.Reader.ReadFilePB(ctx, file)
	default:
		// By default, assume that the user expects to read from the local filesystem.
		return n.Reader.ReadFilePB(ctx, file)
	}
}

func shortFileUrl(url string) string {
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	return url
}
