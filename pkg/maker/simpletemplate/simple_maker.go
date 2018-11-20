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

// Package simpletemplate contains helpers for constructing objects with the
// assumption that cluster objects are simple go templates. That means, that
// the Objects present in the ComponentPackage are treated as go-templates.
package simpletemplate

import (
	"bytes"
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/maker"
)

// Maker makes components via go-templating.
type Maker struct{}

// MakeComponent treats objects in the components as go templates, applying the
// parameters, and then returning the completed objects.
func (m *Maker) MakeComponent(comp *bundle.ComponentPackage, p maker.ParamMaker, of *filter.Options) (*bundle.ComponentPackage, error) {
	return maker.MakeCommon(comp, p, of, makeObject)
}

var _ maker.ComponentMaker = &Maker{}

// detemplatize the objects and return a buffer containing the detemplatized
// YAML.
func makeObject(obj *unstructured.Unstructured, ref bundle.ComponentReference, pm maker.ParamMaker) (*unstructured.Unstructured, error) {
	json, err := converter.FromObject(obj).ToJSON()
	if err != nil {
		return nil, fmt.Errorf("error converting object named %q to json: %v", obj.GetName(), err)
	}

	tmpl, err := template.New(ref.ComponentName + "-tmpl").Parse(string(json))
	if err != nil {
		return nil, fmt.Errorf("error parsing template for component %v: %v", ref, err)
	}

	tmplParams, err := pm()
	if err != nil {
		return nil, fmt.Errorf("error making params for component %v: %v", ref, err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, tmplParams)
	if err != nil {
		return nil, fmt.Errorf("error executing template for component %v: %v", ref, err)
	}

	uns, err := converter.FromJSONString(buf.String()).ToUnstructured()
	if err != nil {
		return nil, err
	}

	return uns, err
}

var _ maker.ObjHandler = makeObject
