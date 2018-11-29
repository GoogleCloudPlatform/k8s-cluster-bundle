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

package simpletemplate

import (
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
)

var component = `
kind: ComponentPackage
spec:
  componentName: logger
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      name: logger-pod
    spec:
      dnsPolicy: '{{.DNSPolicy}}'
      containers:
      - name: logger
        image: '{{.ContainerImage}}'
`

func TestSimpleApplier(t *testing.T) {
	comp, err := converter.FromYAMLString(component).ToComponentPackage()
	if err != nil {
		t.Fatalf("Error converting component to yaml: %v", err)
	}

	usedParams := map[string]interface{}{
		"DNSPolicy":      "FooBarPolicy",
		"ContainerImage": "MyContainerImage",
	}
	notUsedParams := map[string]interface{}{
		"Dapper": "Catamaran",
		"Foo":    "Blarg",
	}

	opts := func() options.JSONOptions {
		allMap := map[string]interface{}{}
		for k, v := range usedParams {
			allMap[k] = v
		}
		for k, v := range notUsedParams {
			allMap[k] = v
		}
		return allMap
	}()

	newComp, err := NewApplier().ApplyOptions(comp, opts)
	if err != nil {
		t.Fatalf("Error applying options: %v", err)
	}
	if newComp == nil {
		t.Fatalf("new-component must not be nil")
	}
	if len(newComp.Spec.Objects) == 0 {
		t.Fatalf("no objects found in new component")
	}

	strval, err := (&converter.ObjectExporter{newComp.Spec.Objects}).ExportAsYAML()
	if err != nil {
		t.Fatalf("Error converting objects to yaml: %v", err)
	}

	for _, val := range usedParams {
		vstr := val.(string)
		if !strings.Contains(strval, vstr) {
			t.Errorf("expected object yaml:\n%s\nto contain %q", strval, vstr)
		}
	}
	for _, val := range notUsedParams {
		vstr := val.(string)
		if strings.Contains(strval, vstr) {
			t.Errorf("expected object yaml:\n%s\nto NOT contain %q", strval, vstr)
		}
	}
}

var multiComponent = `
kind: ComponentPackage
spec:
  componentName: logger
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      name: logger-pod
      annotations:
        floof: indeed
    spec:
      dnsPolicy: '{{.DNSPolicy}}'
      containers:
      - name: logger
        image: '{{.ContainerImage}}'
  - apiVersion: v1
    kind: Pod
    metadata:
      name: dap-pod
    spec:
      containers:
      - name: dapper
        image: gcr.io/floof/{{.DapperImage}}
      - name: verydapper
        image: gcr.io/floof/dapper`

func TestSimpleApplier_MultiItems(t *testing.T) {
	comp, err := converter.FromYAMLString(multiComponent).ToComponentPackage()
	if err != nil {
		t.Fatalf("Error converting component to yaml: %v", err)
	}

	usedParams := map[string]interface{}{
		"DNSPolicy":      "FooBarPolicy",
		"ContainerImage": "MyContainerImage",
		"DapperImage":    "Zorp",
	}
	notUsedParams := map[string]interface{}{
		"Dapper": "Catamaran",
		"Foo":    "Blarg",
	}

	opts := func() options.JSONOptions {
		allMap := map[string]interface{}{}
		for k, v := range usedParams {
			allMap[k] = v
		}
		for k, v := range notUsedParams {
			allMap[k] = v
		}
		return allMap
	}()

	newComp, err := NewApplier().ApplyOptions(comp, opts)
	if err != nil {
		t.Fatalf("Error converting applying options: %v", err)
	}
	if newComp == nil {
		t.Fatalf("new-component must not be nil")
	}
	if len(newComp.Spec.Objects) == 0 {
		t.Fatalf("no objects found in new component")
	}

	strval, err := (&converter.ObjectExporter{newComp.Spec.Objects}).ExportAsYAML()
	if err != nil {
		t.Fatalf("Error converting objects to yaml: %v", err)
	}

	for _, val := range usedParams {
		vstr := val.(string)
		if !strings.Contains(strval, vstr) {
			t.Errorf("expected object yaml:\n%s\nto contain %q", strval, vstr)
		}
	}
	for _, val := range notUsedParams {
		vstr := val.(string)
		if strings.Contains(strval, vstr) {
			t.Errorf("expected object yaml:\n%s\nto NOT contain %q", strval, vstr)
		}
	}
}
