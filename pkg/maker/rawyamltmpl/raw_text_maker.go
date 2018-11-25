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

// Package rawtexttmpl is a special-case component maker for constructing
// objects with the assumption that objects have been inlined via RawText into
// ConfigMaps.
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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/maker"
)

// Maker makes components via go-templating, for objects stored as raw strings
// in config maps.
type Maker struct{}

func (m *Maker) MakeComponent(comp *bundle.ComponentPackage, pm maker.ParamMaker, of *filter.Options) (*bundle.ComponentPackage, error) {
	comp = comp.DeepCopy()
	ref := comp.ComponentReference()

	if len(comp.Spec.Objects) == 0 {
		return nil, fmt.Errorf("no objects found for component %v", ref)
	}

	// Filter to only get the raw text config maps.
	objs := filter.Filter().Objects(comp.Spec.Objects, &filter.Options{
		Annotations: map[string]string{
			string(bundle.InlineTypeIdentifier): string(bundle.RawStringInline),
		},
		Kinds:    []string{"ConfigMap"},
		KeepOnly: true,
	})

	// Construct the objects.
	var newObj []*unstructured.Unstructured
	for _, obj := range objs {
		cfgName := obj.GetName()
		cfgMap := &corev1.ConfigMap{}
		err := converter.FromUnstructured(obj).ToObject(cfgMap)
		if err != nil {
			return nil, fmt.Errorf("error converting from unstructured to config map for object %q", obj.GetName())
		}

		for k, data := range cfgMap.Data {
			fobj, err := makeObject(cfgName, k, data, pm)
			if err != nil {
				return nil, fmt.Errorf("error rendering object with name %q in component %q: %v", obj.GetName(), ref.ComponentName, err)
			}
			newObj = append(newObj, fobj)
		}
	}
	outObj := filter.Filter().Objects(newObj, of)
	comp.Spec.Objects = outObj
	return comp, nil
}

var _ maker.ComponentMaker = &Maker{}

// detemplatize the objects and return the finished unstructured object.
func makeObject(cfgName, key, data string, pm maker.ParamMaker) (*unstructured.Unstructured, error) {
	tmpl, err := template.New(cfgName + "-" + key + "-tmpl").Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing template for config map %q, data key %q: %v", cfgName, key, err)
	}

	tmplParams, err := pm()
	if err != nil {
		return nil, fmt.Errorf("error making params for config map %q, data key %q: %v", cfgName, key, err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, tmplParams)
	if err != nil {
		return nil, fmt.Errorf("error executing template for config map %q, data key %q: %v", cfgName, key, err)
	}

	uns, err := converter.FromJSONString(buf.String()).ToUnstructured()
	if err != nil {
		return nil, err
	}

	return uns, err
}
