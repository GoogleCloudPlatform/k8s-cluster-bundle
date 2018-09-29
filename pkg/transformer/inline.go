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
	"strings"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/golang/protobuf/proto"
)

// Inliner inlines bundle files by reading them from the local or a remote
// filesystem.
type Inliner struct {
	// Local reader reads from the local filesystem.
	Reader core.FilePBReader
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
		Reader: &core.LocalFilePBReader{cwd, &core.LocalFileSystemReader{}},
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
	if err := n.processClusterComponentFiles(ctx, b); err != nil {
		return nil, err
	}

	if err := n.processNodeConfigFiles(ctx, b); err != nil {
		return nil, err
	}

	if opt.TopLayerOnly {
		return b, nil
	}

	// Process all node-bootstrap files.
	for _, v := range spec.GetNodeConfigs() {
		k := v.GetMetadata().GetName()
		if v.GetExternalInitFile() != nil {
			contents, err := n.readFilePB(ctx, v.GetExternalInitFile())
			if err != nil {
				return nil, fmt.Errorf("error processing init script for node bootstrap config %q: %v", k, err)
			}
			v.InitData = &bpb.NodeConfig_InitFile{string(contents)}
		}
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

func (n *Inliner) readFileToProto(ctx context.Context, file *bpb.File, conv *converter.Converter) (proto.Message, error) {
	contents, err := n.readFilePB(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %v", file.GetUrl(), err)
	}
	ext := filepath.Ext(file.GetUrl())
	switch ext {
	case ".yaml":
		return conv.YAMLToProto(contents)
	case ".json":
		return conv.JSONToProto(contents)
	default:
		return conv.YAMLToProto(contents)
	}
}

func (n *Inliner) processClusterComponentFiles(ctx context.Context, b *bpb.ClusterBundle) error {
	for _, cf := range b.GetSpec().GetComponentFiles() {
		pb, err := n.readFileToProto(ctx, cf, converter.ClusterComponent)
		if err != nil {
			return fmt.Errorf("error reading component: %v", err)
		}
		comp := converter.ToClusterComponent(pb)
		compName := comp.GetMetadata().GetName()
		if compName == "" {
			return fmt.Errorf("no component name (metadata.name) found for component with url %q",
				cf.GetUrl())
		}

		// Transform the file-urls of the children to now be relative to where the component is
		for _, ocf := range comp.GetClusterObjectFiles() {
			compUrl := cf.GetUrl()
			objUrl := ocf.GetUrl()
			if strings.HasPrefix(objUrl, "file://") && strings.HasPrefix(compUrl, "file://") {
				ocf.Url = "file://" + filepath.Join(filepath.Dir(shortFileUrl(compUrl)), shortFileUrl(objUrl))
			}
		}

		b.GetSpec().Components = append(b.GetSpec().Components, comp)
	}
	var emptyFiles []*bpb.File
	b.GetSpec().ComponentFiles = emptyFiles
	return nil
}

func (n *Inliner) processNodeConfigFiles(ctx context.Context, b *bpb.ClusterBundle) error {
	for _, cf := range b.GetSpec().GetNodeConfigFiles() {
		pb, err := n.readFileToProto(ctx, cf, converter.NodeConfig)
		if err != nil {
			return fmt.Errorf("error reading node config: %v", err)
		}
		cfg := converter.ToNodeConfig(pb)
		if cfgName := cfg.GetMetadata().GetName(); cfgName == "" {
			return fmt.Errorf("no node config name (metadata.name) found for node config with url %q",
				cf.GetUrl())
		}

		// Transform the init file to be raltive to the node config.
		cfgUrl := cf.GetUrl()
		extInit := cfg.GetExternalInitFile()
		if extInit != nil && strings.HasPrefix(cfgUrl, "file://") && strings.HasPrefix(extInit.GetUrl(), "file://") {
			extInit.Url = "file://" + filepath.Join(filepath.Dir(shortFileUrl(cfgUrl)), shortFileUrl(extInit.GetUrl()))
		}
		b.GetSpec().NodeConfigs = append(b.GetSpec().NodeConfigs, cfg)
	}
	var emptyFiles []*bpb.File
	b.GetSpec().NodeConfigFiles = emptyFiles
	return nil
}

func (n *Inliner) processClusterObjects(ctx context.Context, compName string, b *bpb.ClusterComponent) error {
	for _, cf := range b.GetClusterObjectFiles() {
		pb, err := n.readFileToProto(ctx, cf, converter.Struct)
		if err != nil {
			return fmt.Errorf("error reading component object for component %q: %v", compName, err)
		}
		b.ClusterObjects = append(b.ClusterObjects, converter.ToStruct(pb))
	}
	var emptyFiles []*bpb.File
	b.ClusterObjectFiles = emptyFiles
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
