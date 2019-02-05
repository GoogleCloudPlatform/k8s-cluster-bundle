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
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

const component = `
kind: Component
metadata:
  creationTimestamp: null
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: foo
`

func TestPatchBuilder(t *testing.T) {
	testCases := []struct {
		desc         string
		component    string
		output       string
		opts         map[string]interface{}
		customFilter *filter.Options

		expErrSubstr string
	}{
		{
			desc:      "success: no patch",
			component: component,
			output:    component,
		},
		{
			desc: "success: patch, no options",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplateBuilder
    apiVersion: bundle.gke.io/v1alpha1
    template: |
      kind: Pod
      metadata:
        namespace: foo
`,
			output: `
kind: Component
metadata:
  creationTimestamp: null
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - apiVersion: bundle.gke.io/v1alpha1
    kind: PatchTemplate
    metadata:
      creationTimestamp: null
    template: |
      kind: Pod
      metadata:
        namespace: foo
`,
		},
		{
			desc: "success: patch, build options",
			opts: map[string]interface{}{
				"Namespace": "foo",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplateBuilder
    apiVersion: bundle.gke.io/v1alpha1
    template: |
      kind: Pod
      metadata:
        namespace: {{.Namespace}}
`,
			output: `
kind: Component
metadata:
  creationTimestamp: null
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - apiVersion: bundle.gke.io/v1alpha1
    kind: PatchTemplate
    metadata:
      creationTimestamp: null
    template: |
      kind: Pod
      metadata:
        namespace: foobar
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := converter.FromYAMLString(tc.component).ToComponent()
			if err != nil {
				t.Fatalf("Error converting component %s: %v", tc.component, err)
			}

			hasErr := false
			newComp, err := BuildComponentPatchTemplates(c, tc.customFilter, tc.opts)
			cerr := testutil.CheckErrorCases(err, tc.expErrSubstr)
			if cerr != nil {
				hasErr = true
				t.Error(cerr)
			}
			if err != nil {
				// We hit an expected error, but we can't continue on because newComp is nil.
				return
			}

			compBytes, err := converter.FromObject(newComp).ToYAML()
			if err != nil {
				t.Fatalf("Error converting back to yaml: %v", err)
			}

			compStr := strings.Trim(string(compBytes)," \n\r")
			expStr := strings.Trim(tc.output, " \n\r")
			if expStr != compStr {
				t.Errorf("expected output yaml to be %#v but got %#v", expStr, compStr)
			}
			if hasErr {
				t.Errorf("got yaml contents:\n%s", compStr)
			}
		})
	}
}
