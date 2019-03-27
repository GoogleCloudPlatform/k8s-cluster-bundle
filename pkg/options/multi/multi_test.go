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

package multi

import (
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

func TestMultiApply(t *testing.T) {
	component := `
kind: Component
spec:
  objects:
  - apiVersion: apps/v1beta1
    kind: Deployment
    metadata:
      namespace: foo
  - kind: ObjectTemplate
    type: go-template
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: logger-pod
      spec:
        dnsPolicy: {{.DNSPolicy}}
        containers:
        - name: logger
          image: {{.ContainerImage}}
  - kind: PatchTemplate
    template: |
      kind: Deployment
      metadata:
        namespace: bar
  - kind: PatchTemplate
    template: |
      kind: Pod
      metadata:
        namespace: {{.PodNamespace}}
`

	opts := map[string]interface{}{
		"DNSPolicy":      "somePolicy",
		"ContainerImage": "gcr.io/foobar:latest",
		"PodNamespace":   "zed",
		"Unused":         "unused",
	}

	expStrs := []string{"namespace: zed", "dnsPolicy: somePolicy", "image: gcr.io/foobar:latest", "namespace: bar"}
	notExpStrs := []string{"unused", "namespace: foo", "PatchTemplate", "{{"}

	comp, err := converter.FromYAMLString(component).ToComponent()
	if err != nil {
		t.Fatal(err)
	}

	newComp, err := NewDefaultApplier().ApplyOptions(comp, opts)
	if err != nil {
		t.Fatal(err)
	}

	newCompStr, err := converter.FromObject(newComp).ToYAMLString()
	if err != nil {
		t.Fatal(err)
	}

	hasErr := false
	for _, e := range expStrs {
		if !strings.Contains(newCompStr, e) {
			t.Errorf("expected component to contain string %q", e)
			hasErr = true
		}
	}

	for _, e := range notExpStrs {
		if strings.Contains(newCompStr, e) {
			t.Errorf("expected component to *not* contain string %q", e)
			hasErr = true
		}
	}

	if hasErr {
		t.Errorf("got component\n%s", newCompStr)
	}
}
