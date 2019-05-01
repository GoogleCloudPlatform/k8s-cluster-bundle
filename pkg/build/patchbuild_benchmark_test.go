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

package build

import (
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"

	"testing"
)

func BenchmarkBuildAndPatch_Component(t *testing.B) {
	component := `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplateBuilder
    apiVersion: bundle.gke.io/v1alpha1
    buildSchema:
      required:
        - Namespace
      properties:
        Namespace:
          type: string
    targetSchema:
      required:
      - PodName
      properties:
        PodName:
          type: string
    template: |
      kind: Pod
      metadata:
        namespace: {{.Namespace}}
        name: {{.PodName}}`

	for i := 0; i < t.N; i++ {
		c, err := converter.FromYAMLString(component).ToComponent()
		if err != nil {
			t.Fatal("error parsing component")
		}
		newComp, err := ComponentPatchTemplates(c, &filter.Options{}, map[string]interface{}{
			"Namespace": "foo",
		})
		if err != nil {
			t.Fatal("error patching component")
		}
		_, err = converter.FromObject(newComp).ToYAML()
		if err != nil {
			t.Fatal("error converting object")
		}
	}

}
