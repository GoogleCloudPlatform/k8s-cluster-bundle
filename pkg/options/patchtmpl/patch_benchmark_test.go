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

package patchtmpl

import (
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
)

func BenchmarkPatchTemplateApplier(t *testing.B) {
	component := `kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: derp
  - kind: PatchTemplate
    template: |
      kind: Pod
      metadata:
        namespace: {{.Name}}
  - kind: PatchTemplate
    template: |
      kind: Pod
      metadata:
        annotations:
          fooAnnot: {{.Annot}}
`
	patchOptions := map[string]interface{}{
		"Name":  "doom",
		"Annot": "slayer",
	}
	customFilter := &filter.Options{
		Annotations: map[string]string{
			"phase": "build",
		},
	}

	for i := 0; i < t.N; i++ {
		patcher := NewApplier(DefaultPatcherScheme(), customFilter, true)
		compObj, err := converter.FromYAMLString(component).ToComponent()
		if err != nil {
			t.Fatal(err)
		}
		_, err = patcher.ApplyOptions(compObj, patchOptions)
		if err != nil {
			t.Fatal(err)
		}
	}
}
