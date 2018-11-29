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

// Package inline turns
package inline

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
)

// Inliner inlines data files by reading them from the local or a remote
// filesystem.
type Inliner struct {
	// Readers reads from the local filesystem.
	Readers map[files.URLScheme]files.FileObjReader

	// Rewriters are used for path-rewriting, in the case of relative paths.
	Rewriters map[files.URLScheme]PathRewriter
}

// NewLocalInliner creates a new inliner that only knows how to read local
// files from disk. If the data is stored on disk, the cwd should be the path
// to the directory containing the data file on disk.
func NewLocalInliner(cwd string) *Inliner {
	return NewInlinerWithScheme(
		files.FileScheme,
		&files.LocalFileObjReader{filepath.Dir(cwd), &files.LocalFileSystemReader{}},
		DefaultPathRewriter,
	)
}

// NewInlinerWithScheme creates a new inliner given a URL scheme.
func NewInlinerWithScheme(scheme files.URLScheme, objReader files.FileObjReader, pathRW PathRewriter) *Inliner {
	rdrMap := map[files.URLScheme]files.FileObjReader{
		scheme: objReader,
	}
	rwMap := map[files.URLScheme]PathRewriter{
		scheme: pathRW,
	}
	return &Inliner{
		Readers:   rdrMap,
		Rewriters: rwMap,
	}
}

// InlineBundleFiles converts dereferences file-references in for bundle files and turns
// them into components. Thus, the returned data is a copy with the
// file-references removed.
func (n *Inliner) InlineBundleFiles(ctx context.Context, data *bundle.Bundle) (*bundle.Bundle, error) {
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
		err = n.rewriteObjectPaths(ctx, f, comp)
		if err != nil {
			return nil, fmt.Errorf("error rewriting object paths: %v", err)
		}

		out = append(out, comp)
	}
	newBundle := data.DeepCopy()
	newBundle.Components = out
	newBundle.ComponentFiles = nil
	return newBundle, nil
}

var onlyWhitespace = regexp.MustCompile(`^\s*$`)
var multiDoc = regexp.MustCompile("---(\n|$)")

// InlineComponent reads file-references for component objects.
// The returned components are copies with the file-references removed.
func (n *Inliner) InlineComponent(ctx context.Context, comp *bundle.ComponentPackage) (*bundle.ComponentPackage, error) {
	comp = comp.DeepCopy()
	name := comp.Spec.ComponentName
	var newObjs []*unstructured.Unstructured
	for _, cf := range comp.Spec.ObjectFiles {
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
				return nil, fmt.Errorf("error converting object for component %q in file %q", name, cf.URL)
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

	for _, fg := range comp.Spec.RawTextFiles {
		fgName := fg.Name
		if fgName == "" {
			return nil, fmt.Errorf("error reading raw text file group object for component %q; name was empty ", name)
		}
		m := newConfigMapMaker(fgName)
		for _, cf := range fg.Files {
			text, err := n.readFile(ctx, cf)
			if err != nil {
				return nil, fmt.Errorf("error reading raw text object for component %q: %v", name, err)
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

// InlineComponentsInBundle inlines all the components' objects in a Bundle object.
func (n *Inliner) InlineComponentsInBundle(ctx context.Context, data *bundle.Bundle) (*bundle.Bundle, error) {
	cmp, err := n.InlineAllComponents(ctx, data.Components)
	if err != nil {
		return nil, err
	}
	newb := data.DeepCopy()
	newb.Components = cmp
	return newb, nil
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

// rewriteObjPaths rewrites relative-path'd file-references during component
// inlining.
func (n *Inliner) rewriteObjectPaths(ctx context.Context, compFile bundle.File, comp *bundle.ComponentPackage) error {
	compURL, err := compFile.ParsedURL()
	if err != nil {
		return err
	}

	for i, o := range comp.Spec.ObjectFiles {
		objURL, err := o.ParsedURL()
		if err != nil {
			return err
		}
		scheme := files.URLScheme(objURL.Scheme)
		if scheme == files.EmptyScheme {
			scheme = files.FileScheme
		}
		rw, ok := n.Rewriters[scheme]
		if !ok {
			// It's not really an error to not rewrite paths; Most URL schemes won't
			// provide the ability to rewrite paths.
			continue
		}
		comp.Spec.ObjectFiles[i].URL = rw.RewriteObjectPath(compURL, objURL)
	}

	for i, fg := range comp.Spec.RawTextFiles {
		for j, o := range fg.Files {
			objURL, err := o.ParsedURL()
			if err != nil {
				return err
			}
			scheme := files.URLScheme(objURL.Scheme)
			if scheme == files.EmptyScheme {
				scheme = files.FileScheme
			}
			rw, ok := n.Rewriters[scheme]
			if !ok {
				// It's not really an error to not rewrite paths; Most URL schemes won't
				// provide the ability to rewrite paths.
				continue
			}
			fg.Files[j].URL = rw.RewriteObjectPath(compURL, objURL)
		}
		comp.Spec.RawTextFiles[i] = fg
	}
	return nil
}
