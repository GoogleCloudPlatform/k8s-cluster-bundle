// Copyright 2019 Google LLC
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

// Package gotmpl creates objects from ObjectTemplate objects for
// ObjectTemplates of type "go-template". Once the options are applied to the
// go template, the ObjectTemplate is removed from the component's list of
// objects.
//
// The templates are assumed to be YAML.
package gotmpl

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/openapi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	multiDoc       = regexp.MustCompile("(^|\n)---(\n|$)")
	onlyWhitespace = regexp.MustCompile(`^\s*$`)
)

// applier applies options via go-templating for objects stored as raw strings
// in config maps. This applier only applies to `RawTextFiles` that have been
// inlined as part of the component inlining process.
type applier struct{}

// NewApplier creates a new options applier instance.
func NewApplier() options.Applier {
	return &applier{}
}

// ApplyOptions applies options to the raw-text that's been inlined. To be
// precise, this method looks for ConfigMaps objects with the inline-annotation.
//
// Each key-value pair in the config map is treated as a separate go template
// that represents a single object. Once parameters are applied via the go
// template, the objects are parsed as unstructured objects and added to the
// component's object list. The original ConfigMap is not included in the final
// component.
func (m *applier) ApplyOptions(comp *bundle.Component, opts options.JSONOptions) (*bundle.Component, error) {
	// Make a copy to avoid confusing behavior.
	comp = comp.DeepCopy()

	matched, notMatched := options.PartitionObjectTemplates(comp.Spec.Objects, string(bundle.TemplateTypeGo))

	newObjs, err := options.ApplyCommon(comp.ComponentReference(), matched, opts, applyOptions)
	if err != nil {
		return comp, err
	}
	comp.Spec.Objects = append(notMatched, newObjs...)
	return comp, nil
}

func applyOptions(obj *unstructured.Unstructured, ref bundle.ComponentReference, opts options.JSONOptions) ([]*unstructured.Unstructured, error) {
	// TODO(kashomon): this should probably clone the options for safety.
	objTmpl := &bundle.ObjectTemplate{}
	err := converter.FromUnstructured(obj).ToObject(objTmpl)
	if err != nil {
		return nil, err
	}

	if objTmpl.OptionsSchema != nil {
		opts, err = openapi.ApplyDefaults(opts, objTmpl.OptionsSchema)
		if err != nil {
			return nil, fmt.Errorf("applying schema defaults for object template named %q: %v", obj.GetName(), err)
		}
	}

	tmpl, err := template.New(ref.ComponentName + "-" + obj.GetName() + "-tmpl").Parse(objTmpl.Template)
	if err != nil {
		return nil, fmt.Errorf("error parsing template for object %q: %v", obj.GetName(), err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, opts)
	if err != nil {
		return nil, fmt.Errorf("error executing template for object %v: %v", obj.GetName(), err)
	}

	// TODO(kashomon): It would be better to use YAML-stream decoding, as can be
	// performed by go-yaml: https://godoc.org/gopkg.in/yaml.v2
	var out []*unstructured.Unstructured
	yamlStr := buf.String()
	if multiDoc.MatchString(yamlStr) {
		docs := multiDoc.Split(string(yamlStr), -1)
		for _, doc := range docs {
			if onlyWhitespace.MatchString(doc) {
				continue
			}
			uns, err := converter.FromYAMLString(doc).ToUnstructured()
			if err != nil {
				return nil, err
			}
			out = append(out, uns)
		}
	} else {
		uns, err := converter.FromYAMLString(yamlStr).ToUnstructured()
		if err != nil {
			return nil, err
		}
		out = append(out, uns)
	}

	return out, err
}
