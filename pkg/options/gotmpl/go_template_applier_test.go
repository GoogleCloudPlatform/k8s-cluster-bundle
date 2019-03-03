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

package gotmpl

import (
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
)

var dataComponentMulti = `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.DNSPolicy}}'
        containers:
        - name: logger
          image: '{{.ContainerImage}}'
  - kind: ObjectTemplate
    metadata:
      name: logger-pod-no-inline
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.AnotherDNSPolicy}}'
        containers:
        - name: logger
          image: '{{.AnotherContainerImage}}'
`

func TestRawStringApplier_MultiItems(t *testing.T) {
	var dataComponentBasic = `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.DNSPolicy}}'
        containers:
        - name: logger
          image: '{{.ContainerImage}}'`
	comp, err := converter.FromYAMLString(dataComponentBasic).ToComponent()
	if err != nil {
		t.Fatal(err)
	}

	usedParams := map[string]interface{}{
		"DNSPolicy":      "FooBarPolicy",
		"ContainerImage": "MyContainerImage",
	}
	notUsedParams := map[string]interface{}{
		"AnotherDNSPolicy":      "BlooBlarPolicy",
		"AnotherContainerImage": "AnotherContainerImageVal",
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

	strval, err := (&converter.ObjectExporter{Objects: newComp.Spec.Objects}).ExportAsYAML()
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
