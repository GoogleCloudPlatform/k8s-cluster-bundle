// Copyright 2018-2019 Google LLC
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

package converter

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// Exporter exports cluster objects and components in forms other than just
// basic serialized yaml.
type Exporter struct {
	comps []*bundle.Component
}

// NewExporter creates a new exporter Object
func NewExporter(comps ...*bundle.Component) *Exporter {
	return &Exporter{comps}
}

// ObjectsAsMultiYAML converts cluster objects into multiple YAML files.
func (e *Exporter) ObjectsAsMultiYAML() ([]string, error) {
	var out []string
	var empty []string
	for _, c := range e.comps {
		for _, o := range c.Spec.Objects {
			yaml, err := FromObject(o).ToYAML()
			if err != nil {
				return empty, err
			}
			out = append(out, string(yaml))
		}
	}
	return out, nil
}

// ObjectsAsYAML converts cluster objects into single YAML file.
func (e *Exporter) ObjectsAsSingleYAML() (string, error) {
	var sb strings.Builder
	numComp := len(e.comps)
	for j, c := range e.comps {
		numObj := len(c.Spec.Objects)
		for i, o := range c.Spec.Objects {
			yaml, err := FromObject(o).ToYAML()
			if err != nil {
				return "", err
			}
			sb.Write(yaml)
			if i < numObj-1 || j < numComp-1 {
				// Join the objects into one document.
				// Note: Each doc ends with a newline (from the ToYAML step), so we don't
				// need to write an additional newline
				sb.WriteString("---\n")
			}
		}
	}
	return sb.String(), nil
}

// ComponentSet creates a component set with the given name and version.
func (e *Exporter) ComponentSet(setName, setVersion string) (*bundle.ComponentSet, error) {
	if setName == "" {
		return nil, errors.New("set name is required to export ComponentSet, but was empty")
	}
	if setVersion == "" {
		return nil, errors.New("set version is required to export ComponentSet, but was empty")
	}
	newBun := &bundle.Bundle{
		SetName:    setName,
		Version:    setVersion,
		Components: e.comps,
	}
	return newBun.ComponentSet(), nil
}

// ComponentsAsSingleYAML creates a single YAML file, delimited by '---'. If
// a component set is defined, the component set is append to the end.
func (e *Exporter) ComponentsAsSingleYAML(set *bundle.ComponentSet) (string, error) {
	var sb strings.Builder
	for i, c := range e.comps {
		cyaml, err := FromObject(c).ToYAML()
		if err != nil {
			return "", fmt.Errorf("while rendering component %v: %v", c.ComponentReference(), err)
		}
		sb.Write(cyaml)
		if i < len(e.comps)-1 {
			sb.WriteString("---\n")
		}
	}
	if set != nil {
		sb.WriteString("---\n")
		cyaml, err := FromObject(set).ToYAML()
		if err != nil {
			return "", fmt.Errorf("while rendering component set: %v", err)
		}
		sb.Write(cyaml)
	}
	return sb.String(), nil
}

// ComponentsWithPathTemplates takes in a set of pathTemplates and an optional
// set, and produces a map of paths to YAML files.
//
// pathTemplates are template strings and must be a templated filePath. The
// optional values are
//
// - ComponentName: name of component
// - Version: version of the component or component set
// - BuildTag: build tag for the component
// - SetName: name of the component set
//
// For example, here are same possible ways to craft pathTemplates
//
//   "{{.ComponentName}}/{{.Version}}/{{.ComponentName}}-{{.Version}}.yaml",
//   "{{.ComponentName}}/{{.BuildTag}}/{{.ComponentName}}.yaml",
//   "sets/{{.SetName}}-{{.Version}}.yaml",
func (e *Exporter) ComponentsWithPathTemplates(pathTemplates []string, set *bundle.ComponentSet) (map[string]string, error) {
	var tmpls []*template.Template
	for i, tmplstr := range pathTemplates {
		tmpl, err := template.New(fmt.Sprintf("template-%d", i)).Parse(tmplstr)
		if err != nil {
			return nil, fmt.Errorf("while parsing path template %q: %v", tmplstr, err)
		}
		tmpls = append(tmpls, tmpl)
	}

	outMap := make(map[string]string)
	for _, comp := range e.comps {
		cyaml, err := FromObject(comp).ToYAMLString()
		if err != nil {
			return nil, fmt.Errorf("while rendering component %v: %v", comp.ComponentReference(), err)
		}
		vals := makeValuesForPathTemplates(comp)
		pmap, err := makePathMap(cyaml, tmpls, vals)
		if err != nil {
			return nil, err
		}
		// Merge the values into a single map
		for k, val := range pmap {
			outMap[k] = val
		}
	}

	if set != nil {
		// TODO(kashomon): Could/should sets have BuildTags?
		setVals := []map[string]string{
			map[string]string{
				"SetName": set.Spec.SetName,
				"Version": set.Spec.Version,
			},
		}
		setYaml, err := FromObject(set).ToYAMLString()
		if err != nil {
			return nil, fmt.Errorf("while rendering set: %v", err)
		}
		pmap, err := makePathMap(setYaml, tmpls, setVals)
		if err != nil {
			return nil, err
		}
		for k, val := range pmap {
			outMap[k] = val
		}
	}
	return outMap, nil
}

// makeValuesForPathTemplates creates a map of string-to-string for inserting
// values into the pathTemplates.
//
// Currently it creates a map with three keys.
//
// - ComponentName: name of component
// - Version: version of the component or component set
// - BuildTag: build tag for the component.
//
// If Version is specified, BuildTag will not be specified. Hence, if the
// component has a versiona nd three build tags, four values-maps will be
// generated, one for each of the build tags and one for the version.
//
// Additionally, ComponentName and Version will never be specified if SetName
// and SetVersion are also specified.
func makeValuesForPathTemplates(comp *bundle.Component) []map[string]string {
	values := []map[string]string{
		map[string]string{
			"ComponentName": comp.Spec.ComponentName,
			"Version":       comp.Spec.Version,
		},
	}

	// Create more values if there are any build tags.
	if comp.ObjectMeta.Annotations != nil &&
		comp.ObjectMeta.Annotations[bundle.BuildTagIdentifier] != "" {
		tagString := comp.ObjectMeta.Annotations[bundle.BuildTagIdentifier]
		splitTags := strings.Split(tagString, ",")
		for _, tag := range splitTags {
			values = append(values, map[string]string{
				"ComponentName": comp.Spec.ComponentName,
				"BuildTag":      tag,
				"Version":       comp.Spec.Version,
			})
		}
	}
	return values
}

func makePathMap(comp string, templates []*template.Template, values []map[string]string) (map[string]string, error) {
	pathMap := make(map[string]string)
	var paths []string
	for _, vals := range values {
		for _, pathTmpl := range templates {
			var doc bytes.Buffer
			if err := pathTmpl.Execute(&doc, vals); err != nil {
				return nil, err
			}
			str := doc.String()
			if !strings.Contains(str, "<no value>") {
				paths = append(paths, str)
			}
		}
	}
	for _, p := range paths {
		pathMap[p] = comp
	}
	return pathMap, nil
}
