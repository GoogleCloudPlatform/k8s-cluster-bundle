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

package inline

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
)

// Inliner inlines data files by reading them from the local or a remote
// filesystem.
type Inliner struct {
	// Local reader reads from the local filesystem.
	Reader files.FileObjReader
}

// NewInliner creates a new inliner. If the data is stored on disk, the cwd
// should be the relative path to the directory containing the data file on disk.
func NewLocalInliner(cwd string) *Inliner {
	return &Inliner{
		Reader: &files.LocalFileObjReader{filepath.Dir(cwd), &files.LocalFileSystemReader{}},
	}
}

// Inline converts dereferences file-references in for component data files and
// turns them into components. Thus, the returned data is a copy with the
// file-references removed.
func (n *Inliner) InlineComponentDataFiles(ctx context.Context, data *core.ComponentData) (*core.ComponentData, error) {
	var out []*bundle.ComponentPackage
	for _, f := range data.ComponentFiles {
		contents, err := n.readFile(ctx, f)
		if err != nil {
			return nil, fmt.Errorf("error reading file %q: %v", f.URL, err)
		}
		comp, err := converter.FromFileName(f.URL, contents).ToComponentPackage()
		if err != nil {
			return nil, fmt.Errorf("error converting file %q to a component package: %v", f.URL, err)
		}

		// Because the components can themselves have file references that are
		// relative to the location of the component, we need to transform the
		// references to be based on the location of the component data file.
		compUrl := f.URL
		var newObjFiles []bundle.File
		for _, o := range comp.Spec.ObjectFiles {
			if strings.HasPrefix(o.URL, "file://") && strings.HasPrefix(compUrl, "file://") {
				o.URL = "file://" + filepath.Join(filepath.Dir(shortFileUrl(compUrl)), shortFileUrl(o.URL))
			}
			newObjFiles = append(newObjFiles, o)
		}
		comp.Spec.ObjectFiles = newObjFiles

		// Do the same with the text files.
		var newTextFiles []bundle.File
		for _, o := range comp.Spec.RawTextFiles {
			if strings.HasPrefix(o.URL, "file://") && strings.HasPrefix(compUrl, "file://") {
				o.URL = "file://" + filepath.Join(filepath.Dir(shortFileUrl(compUrl)), shortFileUrl(o.URL))
			}
			newTextFiles = append(newTextFiles, o)
		}
		comp.Spec.RawTextFiles = newTextFiles
		out = append(out, comp)
	}
	return &core.ComponentData{Components: out}, nil
}

var onlyWhitespace = regexp.MustCompile(`^\s*$`)
var multiDoc = regexp.MustCompile("---(\n|$)")

// InlineComponent reads file-references for component objects.
// The returned components are copies with the file-references removed.
func (n *Inliner) InlineComponent(ctx context.Context, comp *bundle.ComponentPackage) (*bundle.ComponentPackage, error) {
	comp = comp.DeepCopy()
	name := comp.Spec.CanonicalName
	var newObjs []*unstructured.Unstructured
	for _, cf := range comp.Spec.ObjectFiles {
		contents, err := n.readFile(ctx, cf)
		if err != nil {
			return nil, fmt.Errorf("error reading file for component %q: %v", name, err)
		}
		ext := filepath.Ext(cf.URL)
		if ext == ".yaml" && multiDoc.Match(contents) {
			splat := multiDoc.Split(string(contents), -1)
			for i, s := range splat {
				if onlyWhitespace.MatchString(s) {
					continue
				}
				obj, err := converter.FromYAMLString(s).ToUnstructured()
				if err != nil {
					return nil, fmt.Errorf("error converting multi-doc object number %d for component %q in file %q", i, name, cf.URL)
				}
				newObjs = append(newObjs, obj)
			}
		} else {
			obj, err := converter.FromFileName(cf.URL, contents).ToUnstructured()
			if err != nil {
				return nil, fmt.Errorf("error converting object for component %q in file %q", name, cf.URL)
			}
			newObjs = append(newObjs, obj)
		}
	}

	for _, cf := range comp.Spec.RawTextFiles {
		text, err := n.readFile(ctx, cf)
		if err != nil {
			return nil, fmt.Errorf("error reading raw text object for component %q: %v", name, err)
		}

		name := filepath.Base(cf.URL)
		m := newConfigMapMaker(name)
		m.addData(name, string(text))
		uns, err := m.toUnstructured()
		if err != nil {
			return nil, fmt.Errorf("error converting text object to unstructured for component %q: %v", name, err)
		}
		newObjs = append(newObjs, uns)
	}
	comp.Spec.RawTextFiles = nil
	comp.Spec.ObjectFiles = nil
	comp.Spec.Objects = newObjs

	return comp, nil
}

// InlineAllComponents inlines objects into ComponentPackages.
func (n *Inliner) InlineAllComponents(ctx context.Context, packs []*bundle.ComponentPackage) ([]*bundle.ComponentPackage, error) {
	var out []*bundle.ComponentPackage
	for _, p := range packs {
		newp, err := n.InlineComponent(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("error in InlineAllComponents: %v", err)
		}
		out = append(out, newp)
	}
	return out, nil
}

// InlineComponentsInData inlines all the components' objects in a ComponentData object.
func (n *Inliner) InlineComponentsInData(ctx context.Context, data *core.ComponentData) (*core.ComponentData, error) {
	cmp, err := n.InlineAllComponents(ctx, data.Components)
	if err != nil {
		return nil, err
	}
	return &core.ComponentData{
		Components: cmp,
	}, nil
}

// readFile from either a local or remote location.
func (n *Inliner) readFile(ctx context.Context, file bundle.File) ([]byte, error) {
	url := file.URL
	if url == "" {
		return nil, fmt.Errorf("file %v was specified but no file path/url was provided", file)
	}
	switch {
	case strings.HasPrefix("gs://", url):
		return nil, fmt.Errorf("url-type (GCS) not supported; file was %q", url)
	case strings.HasPrefix("https://", url) || strings.HasPrefix("http://", url):
		return nil, fmt.Errorf("url-type (HTTP[S]) not supported; file was %q", url)
	case strings.HasPrefix("file://", url):
		return n.Reader.ReadFileObj(ctx, file)
	default:
		// By default, assume that the user expects to read from the local filesystem.
		return n.Reader.ReadFileObj(ctx, file)
	}
}

func shortFileUrl(url string) string {
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	return url
}
