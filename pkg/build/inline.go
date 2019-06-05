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
	"net/url"
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
		&files.LocalFileObjReader{
			WorkingDir: filepath.Dir(cwd),
			Rdr:        &files.LocalFileSystemReader{},
		},
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

// BundleFiles inlines file-references in for bundle files. If
// the bundlePath is defined and not absolute and the scheme is file based
// scheme, then the path is made absolute before proceeding.
func (n *Inliner) BundleFiles(ctx context.Context, data *bundle.BundleBuilder, bundlePath string) (*bundle.Bundle, error) {
	bundleURL, err := url.Parse(bundlePath)
	if err != nil {
		return nil, err
	}
	bundleURL, err = makeAbsForFileScheme(bundleURL)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(bundleURL.Path) {
		return nil, fmt.Errorf("bundlePath must be absolute but was %s", bundleURL.Path)
	}
	var comps []*bundle.Component
	for _, f := range data.ComponentFiles {
		furl, err := f.ParsedURL()
		if err != nil {
			return nil, err
		}
		f.URL = makeAbsWithParent(bundleURL, furl).String()

		contents, err := n.readFile(ctx, f)
		if err != nil {
			return nil, fmt.Errorf("error reading file %q: %v", f.URL, err)
		}
		uns, err := converter.FromFileName(f.URL, contents).ToUnstructured()
		if err != nil {
			return nil, err
		}

		kind := uns.GetKind()
		switch kind {
		case "Component":
			c, err := converter.FromFileName(f.URL, contents).ToComponent()
			if err != nil {
				return nil, err
			}
			comps = append(comps, c)
		case "ComponentBuilder":
			c, err := converter.FromFileName(f.URL, contents).ToComponentBuilder()
			if err != nil {
				return nil, err
			}
			if c.GetName() == "" && data.ComponentNamePolicy == "SetAndComponent" {
				c.ObjectMeta.Name = strings.Join([]string{data.SetName, data.Version, c.ComponentName, c.Version}, "-")
			}
			comp, err := n.ComponentFiles(ctx, c, f.URL)
			if err != nil {
				return nil, err
			}
			comps = append(comps, comp)
		default:
			return nil, fmt.Errorf("unsupported kind for component: %q; only supported kinds are Component and ComponentBuilder", kind)
		}
	}

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
var multiDoc = regexp.MustCompile("(^|\n)---")
var nonDNS = regexp.MustCompile(`[^-a-z0-9\.]`)

// ComponentFiles reads file-references for component builder objects.  The
// returned components are copies with the file-references removed. If the
// componentPath is not absolute and the scheme is a file scheme, it will be
// made absolute before proceeding.
func (n *Inliner) ComponentFiles(ctx context.Context, comp *bundle.ComponentBuilder, componentPath string) (*bundle.Component, error) {
	componentURL, err := url.Parse(componentPath)
	if err != nil {
		return nil, err
	}
	componentURL, err = makeAbsForFileScheme(componentURL)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(componentURL.Path) {
		return nil, fmt.Errorf("componentURL must be absolute but was %s", componentURL.Path)
	}

	newObjs, tmplBuilders, err := n.objectFiles(ctx, comp.ObjectFiles, comp.ComponentReference(), componentURL)
	if err != nil {
		return nil, err
	}

	tmplObjs, err := n.objectTemplateBuilders(ctx, tmplBuilders, comp.ComponentReference())
	if err != nil {
		return nil, err
	}
	newObjs = append(newObjs, tmplObjs...)

	cfgObj, err := n.rawTextFiles(ctx, comp.RawTextFiles, comp.ComponentReference(), componentURL)
	if err != nil {
		return nil, err
	}
	newObjs = append(newObjs, cfgObj...)

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
			Objects:       newObjs,
		},
	}
	return newComp, nil
}

// AllComponentFiles is a convenience method for inlining multiple component files.
func (n *Inliner) AllComponentFiles(ctx context.Context, cbs []*bundle.ComponentBuilder) ([]*bundle.Component, error) {
	var out []*bundle.Component
	for _, cb := range cbs {
		newc, err := n.ComponentFiles(ctx, cb, "")
		if err != nil {
			return nil, err
		}
		out = append(out, newc)
	}
	return out, nil
}

// objectFiles inlines object files. in the success case, it returns
//
// 1.) The inlined object files.
// 2.) A map of path-to-ObjectTemplateBuilder
func (n *Inliner) objectFiles(ctx context.Context, objFiles []bundle.File, ref bundle.ComponentReference, componentPath *url.URL) ([]*unstructured.Unstructured, map[string][]*unstructured.Unstructured, error) {
	var newObjs []*unstructured.Unstructured
	objTmplBuilders := make(map[string][]*unstructured.Unstructured)
	for _, cf := range objFiles {
		furl, err := cf.ParsedURL()
		if err != nil {
			return nil, nil, err
		}
		cf.URL = makeAbsWithParent(componentPath, furl).String()

		contents, err := n.readFile(ctx, cf)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading file %v for component %v: %v", cf, ref, err)
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
					return nil, nil, fmt.Errorf("converting multi-doc object number %d for component %v, %v", i, ref, err)
				}
				if obj.GetKind() == "ObjectTemplateBuilder" {
					if objTmplBuilders[cf.URL] == nil {
						objTmplBuilders[cf.URL] = []*unstructured.Unstructured{obj}
					} else {
						objTmplBuilders[cf.URL] = append(objTmplBuilders[cf.URL], obj)
					}
				} else {
					newObjs = append(newObjs, obj)
				}
			}
		} else {
			obj, err := converter.FromFileName(cf.URL, contents).ToUnstructured()
			if err != nil {
				return nil, nil, fmt.Errorf("for component %q, %v", ref, err)
			}
			if obj.GetKind() == "ObjectTemplateBuilder" {
				objTmplBuilders[cf.URL] = []*unstructured.Unstructured{obj}
			} else {
				newObjs = append(newObjs, obj)
			}
		}
	}
	return newObjs, objTmplBuilders, nil
}

// objectTemplateBuilders builds ObjectTemplates from ObjectTemplateBuilders
func (n *Inliner) objectTemplateBuilders(ctx context.Context, objects map[string][]*unstructured.Unstructured, ref bundle.ComponentReference) ([]*unstructured.Unstructured, error) {
	var outObj []*unstructured.Unstructured
	for parentPath, objList := range objects {
		for _, obj := range objList {
			if obj.GetKind() != "ObjectTemplateBuilder" {
				// There shouldn't be any ObjectTemplateBuilders at this point
				continue
			}
			name := obj.GetName()

			builder := &bundle.ObjectTemplateBuilder{}
			if err := converter.FromUnstructured(obj).ToObject(builder); err != nil {
				return nil, fmt.Errorf("for component %v and object %q: %v", ref, name, err)
			}

			parentURL, err := url.Parse(parentPath)
			if err != nil {
				return nil, err
			}

			furl, err := builder.File.ParsedURL()
			if err != nil {
				return nil, err
			}

			builder.File.URL = makeAbsWithParent(parentURL, furl).String()

			contents, err := n.readFile(ctx, builder.File)
			if err != nil {
				return nil, fmt.Errorf("for component %v and object %q: %v", ref, name, err)
			}

			objTemplate := &bundle.ObjectTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "bundle.gke.io/v1alpha1",
					Kind:       "ObjectTemplate",
				},
				ObjectMeta:    builder.ObjectMeta,
				OptionsSchema: builder.OptionsSchema,
				Template:      string(contents),
			}
			objTemplate.ObjectMeta.Annotations = make(map[string]string)
			for key, value := range builder.ObjectMeta.Annotations {
				objTemplate.ObjectMeta.Annotations[key] = value
			}
			objTemplate.ObjectMeta.Annotations[string(bundle.InlinePathIdentifier)] = builder.File.URL

			tmplType := bundle.TemplateTypeGo
			if builder.Type != bundle.TemplateTypeUndefined {
				tmplType = builder.Type
			}
			objTemplate.Type = tmplType

			objJSON, err := converter.FromObject(objTemplate).ToJSON()
			if err != nil {
				return nil, fmt.Errorf("for component %v and object %q, while converting back to JSON: %v", ref, name, err)
			}

			unsObj, err := converter.FromJSON(objJSON).ToUnstructured()
			if err != nil {
				return nil, fmt.Errorf("for component %v and object %q, while converting back to Unstructured: %v", ref, name, err)
			}
			outObj = append(outObj, unsObj)
		}
	}
	return outObj, nil
}

func (n *Inliner) rawTextFiles(ctx context.Context, fileGroups []bundle.FileGroup, ref bundle.ComponentReference, componentPath *url.URL) ([]*unstructured.Unstructured, error) {
	var newObjs []*unstructured.Unstructured
	for _, fg := range fileGroups {
		fgName := fg.Name
		if fgName == "" {
			return nil, fmt.Errorf("error reading raw text file group object for component %v; name was empty ", ref)
		}
		m := newConfigMapMaker(fgName)
		for _, cf := range fg.Files {
			furl, err := cf.ParsedURL()
			if err != nil {
				return nil, err
			}
			cf.URL = makeAbsWithParent(componentPath, furl).String()

			text, err := n.readFile(ctx, cf)
			if err != nil {
				return nil, fmt.Errorf("error reading raw text file for component %q: %v", ref, err)
			}
			dataName := filepath.Base(cf.URL)
			if fg.AsBinary {
				m.addBinaryData(dataName, text)
			} else {
				m.addData(dataName, string(text))
			}
		}
		if len(m.cfgMap.Data) > 0 && len(m.cfgMap.BinaryData) > 0 {
			return nil, fmt.Errorf("both and binary data were filled out for group: %v", fg)
		}

		for key, value := range fg.Annotations {
			m.cfgMap.ObjectMeta.Annotations[key] = value
		}
		for key, value := range fg.Labels {
			m.cfgMap.ObjectMeta.Labels[key] = value
		}

		uns, err := m.toUnstructured()
		if err != nil {
			return nil, fmt.Errorf("for component %v and file group %q, %v", ref, fgName, err)
		}
		newObjs = append(newObjs, uns)
	}
	return newObjs, nil
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
