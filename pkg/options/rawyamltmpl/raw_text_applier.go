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

// Package rawtexttmpl is a special-case option applier for adding options to
// objects with the assumption that objects have been inlined as go-templates
// via RawTextFiles ComponentSpec field into ConfigMaps.
//
// Once parameters are applied via the go template, the objects are parsed as
// unstructured objects and added to the component's object list. The original
// ConfigMap is not included in the final component.
package rawtexttmpl

import (
	"bytes"
	"fmt"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
)

// applier applies options via go-templating for objects stored as raw strings
// in config maps. This applier only applies to `RawTextFiles` that have been
// inlined as part of the component inlining process.
type applier struct{}

// NewApplier creates a new optionts applier instance.
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
	comp = comp.DeepCopy()
	ref := comp.ComponentReference()

	// Filter to only get the raw text config maps.
	objs := filter.NewFilter().Objects(comp.Spec.Objects, &filter.Options{
		Annotations: map[string]string{
			string(bundle.InlineTypeIdentifier): string(bundle.RawStringInline),
		},
		Kinds:    []string{"ConfigMap"},
		KeepOnly: true,
	})

	// Filter to get the non raw-text objects.
	otherObjs := filter.NewFilter().Objects(comp.Spec.Objects, &filter.Options{
		Annotations: map[string]string{
			string(bundle.InlineTypeIdentifier): string(bundle.RawStringInline),
		},
		Kinds: []string{"ConfigMap"},
	})

	// Construct the objects. We can't use the common applier because each
	// configmap can have multiple go templates, each of which represent a k8s
	// object that needs parameters
	var newObj []*unstructured.Unstructured
	for _, obj := range objs {
		cfgName := obj.GetName()
		cfgMap := &corev1.ConfigMap{}
		err := converter.FromUnstructured(obj).ToObject(cfgMap)
		if err != nil {
			return nil, fmt.Errorf("error converting from unstructured to config map for object %q", obj.GetName())
		}

		for k, data := range cfgMap.Data {
			fobj, err := applyOptions(cfgName, k, data, opts)
			if err != nil {
				return nil, fmt.Errorf("error rendering object with name %q in component %q: %v", obj.GetName(), ref.ComponentName, err)
			}
			newObj = append(newObj, fobj)
		}
	}

	comp.Spec.Objects = append(otherObjs, newObj...)
	return comp, nil
}

func applyOptions(cfgName, key, data string, opts options.JSONOptions) (*unstructured.Unstructured, error) {
	tmpl, err := template.New(cfgName + "-" + key + "-tmpl").Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing template for config map %q, data key %q: %v", cfgName, key, err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, opts)
	if err != nil {
		return nil, fmt.Errorf("error executing template for config map %q, data key %q: %v", cfgName, key, err)
	}

	uns, err := converter.FromJSONString(buf.String()).ToUnstructured()
	if err != nil {
		return nil, err
	}

	return uns, err
}
