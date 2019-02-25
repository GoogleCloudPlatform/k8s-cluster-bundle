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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

var dataComponentMulti = `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    type: go-template
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
    type: go-template
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

func TestGoTemplateApplier(t *testing.T) {
	testCases := []struct {
		desc          string
		component     string
		usedParams    map[string]interface{}
		notUsedParams map[string]interface{}
		expSubstrings []string
		expErrSubstr  string
	}{
		{
			desc: "success",
			component: `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    type: go-template
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.DNSPolicy}}'
        containers:
        - name: logger
          image: '{{.ContainerImage}}'`,
			usedParams: map[string]interface{}{
				"DNSPolicy":      "FooBarPolicy",
				"ContainerImage": "MyContainerImage",
			},
			notUsedParams: map[string]interface{}{
				"AnotherDNSPolicy":      "BlooBlarPolicy",
				"AnotherContainerImage": "AnotherContainerImageVal",
			},
		},
		{
			desc: "multiple object templates",
			component: `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    type: go-template
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
    type: go-template
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.AnotherDNSPolicy}}'
        containers:
        - name: logger
          image: '{{.AnotherContainerImage}}'`,
			usedParams: map[string]interface{}{
				"DNSPolicy":             "FooBarPolicy",
				"ContainerImage":        "MyContainerImage",
				"AnotherDNSPolicy":      "BlooBlarPolicy",
				"AnotherContainerImage": "AnotherContainerImageVal",
			},
			notUsedParams: map[string]interface{}{
				"Foof": "Boof",
			},
		},
		{
			desc: "multi-doc object templates",
			component: `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    type: go-template
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
      ---
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.AnotherDNSPolicy}}'
        containers:
        - name: logger
          image: '{{.AnotherContainerImage}}'`,
			usedParams: map[string]interface{}{
				"DNSPolicy":             "FooBarPolicy",
				"ContainerImage":        "MyContainerImage",
				"AnotherDNSPolicy":      "BlooBlarPolicy",
				"AnotherContainerImage": "AnotherContainerImageVal",
			},
			notUsedParams: map[string]interface{}{
				"Foof": "Boof",
			},
		},
		{
			desc: "object templates of wrong kind",
			component: `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    type: zed
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.DNSPolicy}}'
        containers:
        - name: logger
          image: '{{.ContainerImage}}'`,
			notUsedParams: map[string]interface{}{
				"DNSPolicy":      "FooBarPolicy",
				"ContainerImage": "MyContainerImage",
			},
			expSubstrings: []string{
				"ObjectTemplate",
			},
		},
		{
			desc: "object templates and other objects",
			component: `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: Pod
    metadata:
      name: derp
  - kind: ObjectTemplate
    type: go-template
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.DNSPolicy}}'
        containers:
        - name: logger
          image: '{{.ContainerImage}}'`,
			usedParams: map[string]interface{}{
				"DNSPolicy":      "FooBarPolicy",
				"ContainerImage": "MyContainerImage",
			},
			expSubstrings: []string{
				"derp",
			},
		},
		{
			desc: "object templates: options schema defaulting",
			component: `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    type: go-template
    optionsSchema:
      properties:
        ContainerImage:
          default: 'my-container-image'
          type: string
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.DNSPolicy}}'
        containers:
        - name: logger
          image: '{{.ContainerImage}}'`,
			usedParams: map[string]interface{}{
				"DNSPolicy": "FooBarPolicy",
			},
			expSubstrings: []string{
				"my-container-image",
			},
		},
		{
			desc: "object templates: param validation fail",
			component: `
kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    type: go-template
    optionsSchema:
      properties:
        ContainerImage:
          type: string
          pattern: '^[a-z]+$'
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: '{{.DNSPolicy}}'
        containers:
        - name: logger
          image: '{{.ContainerImage}}'`,
			usedParams: map[string]interface{}{
				"DNSPolicy":      "FooBarPolicy",
				"ContainerImage": "MyContainerImage",
			},
			expErrSubstr: "ContainerImage in body should match",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			comp, err := converter.FromYAMLString(tc.component).ToComponent()
			if err != nil {
				t.Fatal(err)
			}

			opts := func() options.JSONOptions {
				allMap := map[string]interface{}{}
				for k, v := range tc.usedParams {
					allMap[k] = v
				}
				for k, v := range tc.notUsedParams {
					allMap[k] = v
				}
				return allMap
			}()

			newComp, err := NewApplier().ApplyOptions(comp, opts)
			cerr := testutil.CheckErrorCases(err, tc.expErrSubstr)
			if cerr != nil {
				t.Fatal(cerr)
			}
			if err != nil {
				// Even an expected error is terminla
				return
			}
			if newComp == nil {
				t.Fatalf("new-component must not be nil")
			}
			if len(newComp.Spec.Objects) == 0 {
				t.Fatalf("no objects found in new component")
			}

			strval, err := converter.NewExporter(newComp).ObjectsAsSingleYAML()
			if err != nil {
				t.Fatalf("Error converting objects to yaml: %v", err)
			}

			for _, val := range tc.usedParams {
				vstr := val.(string)
				if !strings.Contains(strval, vstr) {
					t.Errorf("got object yaml:\n%s\nbut expected it to contain %q", strval, vstr)
				}
			}
			for _, val := range tc.notUsedParams {
				vstr := val.(string)
				if strings.Contains(strval, vstr) {
					t.Errorf("got object yaml:\n%s\nbut expected it to NOT contain %q", strval, vstr)
				}
			}
			for _, substr := range tc.expSubstrings {
				if !strings.Contains(strval, substr) {
					t.Errorf("got object yaml:\n%s\nbut expected it to contain %q", strval, substr)
				}
			}
		})
	}
}

func TestCopyComponent(t *testing.T) {
	compStr := `kind: Component
spec:
  componentName: data-component
  objects:
  - kind: ObjectTemplate
    type: go-template
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

	comp, err := converter.FromYAMLString(compStr).ToComponent()
	if err != nil {
		t.Fatal(err)
	}

	usedParams := map[string]interface{}{
		"DNSPolicy":      "FooBarPolicy",
		"ContainerImage": "MyContainerImage",
	}

	newComp, err := NewApplier().ApplyOptions(comp, usedParams)
	if err != nil {
		t.Fatal(err)
	}

	origStr, err := converter.FromObject(comp).ToYAMLString()
	if err != nil {
		t.Fatal(err)
	}

	newCompStr, err := converter.FromObject(newComp).ToYAMLString()
	if err != nil {
		t.Fatal(err)
	}

	if newCompStr == origStr {
		t.Fatalf("Got equal component strings, but expected them to be different because the component should bo copied. "+
			"Values were new string:\n%s\nOriginal string:\n%s", newCompStr, origStr)
	}
}
