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
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

func BenchmarkGoTemplateApplier(t *testing.B){
  component := `kind: Component
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
          image: '{{.AnotherContainerImage}}'`

  templates := map[string]interface{}{
    "DNSPolicy": "FooBarPolicy",
    "ContainerImage":        "MyContainerImage",
    "AnotherDNSPolicy":      "BlooBlarPolicy",
    "AnotherContainerImage": "AnotherContainerImageVal",
  }

  for i := 0; i < t.N; i++ {
    comp, err := converter.FromYAMLString(component).ToComponent()
    if err != nil {
      t.Fatal(err)
    }

    _, err = NewApplier().ApplyOptions(comp, templates)
    if err != nil {
      t.Fatal(err)
    }  
  }
  
}
