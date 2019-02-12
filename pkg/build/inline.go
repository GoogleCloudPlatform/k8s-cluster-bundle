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

package build

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Inliner inlines data files by reading them from the local or a remote
// filesystem.
type Inliner struct {
	// Readers reads from the local filesystem.
	Readers map[files.URLScheme]files.FileObjReader
}

// NewLocalInliner creates a new inliner that only knows how to read local
// files from disk. If the data is stored on disk, the cwd should be the path
// to the directory containing the data file on disk. Relative paths are not
// supported.
func NewLocalInliner(cwd string) *Inliner {
	return NewInlinerWithScheme(
		files.FileScheme,
		&files.LocalFileObjReader{filepath.Dir(cwd), &files.LocalFileSystemReader{}},
	)
}

// NewInlinerWithScheme creates a new inliner given a URL scheme.
func NewInlinerWithScheme(scheme files.URLScheme, objReader files.FileObjReader) *Inliner {
	rdrMap := map[files.URLScheme]files.FileObjReader{
		scheme: objReader,
	}
	return &Inliner{
		Readers: rdrMap,
	}
}

// BundleFiles converts dereferences file-references in for bundle files.
func (n *Inliner) BundleFiles(ctx context.Context, data *bundle.BundleBuilder) (*bundle.Bundle, error) {
	var compbs []*bundle.ComponentBuilder
	var comps []*bundle.Component
	for _, f := range data.ComponentFiles {
		contents, err := n.readFile(ctx, f)
		if err != nil {
			return nil, fmt.Errorf("error reading file %q: %v", f.URL, err)
		}
		uns, err := converter.FromFileName(f.URL, contents).ToUnstructured()
		if err != nil {
			return nil, fmt.Errorf("error converting file %q to Unstructured: %v", f.URL, err)
		}

		kind := uns.GetKind()
		switch kind {
		case "Component":
			c, err := converter.FromFileName(f.URL, contents).ToComponent()
			if err != nil {
				return nil, fmt.Errorf("error converting file %q to a component: %v", f.URL, err)
			}
			comps = append(comps, c)
		case "ComponentBuilder":
			c, err := converter.FromFileName(f.URL, contents).ToComponentBuilder()
			if err != nil {
				return nil, fmt.Errorf("error converting file %q to a component builder: %v", f.URL, err)
			}
			if c.GetName() == "" {
				c.ObjectMeta.Name = strings.Join([]string{data.SetName, data.Version, c.ComponentName, c.Version}, "-")
			}
			compbs = append(compbs, c)
		default:
			return nil, fmt.Errorf("unsupported kind for component: %q; only supported kinds are Component and ComponentBuilder", kind)
		}
	}

	inlComps, err := n.AllComponentFiles(ctx, compbs)
	if err != nil {
		return nil, err
	}
	comps = append(comps, inlComps...)

	newBundle := &bundle.Bundle{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "bundle.gke.io/v1alpha1",
			Kind:       "Bundle",
		},
		ObjectMeta: *data.ObjectMeta.DeepCopy(),
		SetName:    data.SetName,
		Version:    data.Version,
		Components: comps,
	}
	return newBundle, nil
}

var onlyWhitespace = regexp.MustCompile(`^\s*$`)
var multiDoc = regexp.MustCompile("---(\n|$)")
var nonDNS = regexp.MustCompile(`[^-a-z0-9\.]`)

// ComponentFiles reads file-references for component builder objects.
// The returned components are copies with the file-references removed.
func (n *Inliner) ComponentFiles(ctx context.Context, comp *bundle.ComponentBuilder) (*bundle.Component, error) {
	name := comp.ComponentName
	var newObjs []*unstructured.Unstructured
	for _, cf := range comp.ObjectFiles {
		contents, err := n.readFile(ctx, cf)
		if err != nil {
			return nil, fmt.Errorf("error reading file %v for component %q: %v", cf, name, err)
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
				annot := obj.GetAnnotations()
				if annot == nil {
					annot = make(map[string]string)
				}
				annot[string(bundle.InlineTypeIdentifier)] = string(bundle.KubeObjectInline)
				obj.SetAnnotations(annot)
				newObjs = append(newObjs, obj)
			}
		} else {
			obj, err := converter.FromFileName(cf.URL, contents).ToUnstructured()
			if err != nil {
				return nil, fmt.Errorf("error converting object to unstructured for component %q in file %q", name, cf.URL)
			}
			annot := obj.GetAnnotations()
			if annot == nil {
				annot = make(map[string]string)
			}
			annot[string(bundle.InlineTypeIdentifier)] = string(bundle.KubeObjectInline)
			obj.SetAnnotations(annot)
			newObjs = append(newObjs, obj)
		}
	}

	for _, fg := range comp.RawTextFiles {
		fgName := fg.Name
		if fgName == "" {
			return nil, fmt.Errorf("error reading raw text file group object for component %q; name was empty ", name)
		}
		m := newConfigMapMaker(fgName)
		for _, cf := range fg.Files {
			text, err := n.readFile(ctx, cf)
			if err != nil {
				return nil, fmt.Errorf("error reading raw text file for component %q: %v", name, err)
			}
			dataName := filepath.Base(cf.URL)
			m.addData(dataName, string(text))
		}
		m.cfgMap.ObjectMeta.Annotations[string(bundle.InlineTypeIdentifier)] = string(bundle.RawStringInline)
		uns, err := m.toUnstructured()
		if err != nil {
			return nil, fmt.Errorf("error converting text object to unstructured for component %q and file group %q: %v", name, fgName, err)
		}
		newObjs = append(newObjs, uns)
	}

	om := *comp.ObjectMeta.DeepCopy()
	if om.Name == "" {
		name := strings.ToLower(comp.ComponentName + `-` + comp.Version)
		om.Name = nonDNS.ReplaceAllLiteralString(name, `-`)
	}
	errs := validation.IsDNS1123Subdomain(om.Name)
	if len(errs) > 0 {
		return nil, fmt.Errorf("metadata.Name %q is not a valid DNS 1123 subdomain in component %q/%q: %v",
			om.Name, comp.ComponentName, comp.Version, errs)
	}
	newComp := &bundle.Component{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "bundle.gke.io/v1alpha1",
			Kind:       "Component",
		},
		ObjectMeta: om,
		Spec: bundle.ComponentSpec{
			ComponentName: comp.ComponentName,
			Version:       comp.Version,
			AppVersion:    comp.AppVersion,
			Objects:       newObjs,
		},
	}
	return newComp, nil
}

// AllComponentFiles is a convenience method for inlining multiple component files.
func (n *Inliner) AllComponentFiles(ctx context.Context, cbs []*bundle.ComponentBuilder) ([]*bundle.Component, error) {
	var out []*bundle.Component
	for _, cb := range cbs {
		newc, err := n.ComponentFiles(ctx, cb)
		if err != nil {
			return nil, err
		}
		out = append(out, newc)
	}
	return out, nil
}

// readFile from either a local or remote location.
func (n *Inliner) readFile(ctx context.Context, file bundle.File) ([]byte, error) {
	parsed, err := file.ParsedURL()
	if err != nil {
		return nil, err
	}
	scheme := files.URLScheme(parsed.Scheme)
	if scheme == files.EmptyScheme {
		scheme = files.FileScheme
	}
	rdr, ok := n.Readers[scheme]
	if !ok {
		return nil, fmt.Errorf("could not find file reader for scheme %q for url %q", parsed.Scheme, file.URL)
	}
	return rdr.ReadFileObj(ctx, file)
}
