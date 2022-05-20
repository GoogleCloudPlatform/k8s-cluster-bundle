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

package patchtmpl

import (
	"errors"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

func TestPatch(t *testing.T) {
	testCases := []struct {
		desc            string
		component       string
		opts            map[string]interface{}
		customFilter    *filter.Options
		removeTemplates bool

		expMatchSubstrs   []string
		expNoMatchSubstrs []string
		expErrSubstr      string
	}{
		{
			desc: "success: no patch",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: foo
`,
			expMatchSubstrs: []string{"namespace: foo"},
		},
		{
			desc: "success: no patch, with remove",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: foo
`,
			expMatchSubstrs: []string{"namespace: foo"},
			removeTemplates: true,
		},
		{
			desc: "success: patch, no options",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    template: |
      metadata:
        namespace: foo`,
			expMatchSubstrs: []string{"metadata:\n      namespace: foo"},
			removeTemplates: true,
		},
		{
			desc: "success: patch, no options, remove",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    template: |
      kind: Pod
      metadata:
        namespace: foo`,
			removeTemplates:   true,
			expMatchSubstrs:   []string{"metadata:\n      namespace: foo"},
			expNoMatchSubstrs: []string{"PatchTemplate"},
		},
		{
			desc: "success: patch, basic options",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    template: |
      kind: Pod
      metadata:
        namespace: {{.Name}}`,
			expMatchSubstrs: []string{"namespace: zed"},
		},
		{
			desc: "success: patch, basic options with default",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    optionsSchema:
      properties:
        Name:
          type: string
          default: zed
    template: |
      kind: Pod
      metadata:
        namespace: {{.Name}}
`,
			expMatchSubstrs: []string{"namespace: zed"},
		},
		{
			desc: "success: patch, basic options, two object-matches",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: derp
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: derpper
  - kind: PatchTemplate
    template: |
      metadata:
        namespace: {{.Name}}
`,
			expMatchSubstrs:   []string{"namespace: zed"},
			expNoMatchSubstrs: []string{"namespace: derp", "namespace: derpper"},
		},
		{
			desc: "success: patch, two objects, one match: kind",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: dorp
  - apiVersion: apps/v1beta1
    kind: Deployment
    metadata:
      namespace: derpper
  - kind: PatchTemplate
    template: |
      metadata:
        namespace: {{.Name}}
    selector:
      kinds:
      - Pod
`,
			expMatchSubstrs:   []string{"namespace: zed", "namespace: derpper"},
			expNoMatchSubstrs: []string{"namespace: dorp"},
		},
		{
			desc: "success: patch, two objects, one match: kind from Template",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: dorp
  - apiVersion: apps/v1beta1
    kind: Deployment
    metadata:
      namespace: derpper
  - kind: PatchTemplate
    template: |
      kind: Pod
      metadata:
        namespace: {{.Name}}
`,
			expMatchSubstrs:   []string{"namespace: zed", "namespace: derpper"},
			expNoMatchSubstrs: []string{"namespace: dorp"},
		},
		{
			desc: "success: patch, two objects, one match: qualified kind from Template",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: dorp
  - apiVersion: apps/v1beta1
    kind: Deployment
    metadata:
      namespace: derpper
  - kind: PatchTemplate
    template: |
      apiVersion: v1
      kind: Pod
      metadata:
        namespace: {{.Name}}
`,
			expMatchSubstrs:   []string{"namespace: zed", "namespace: derpper"},
			expNoMatchSubstrs: []string{"namespace: dorp"},
		},
		{
			desc: "success: patch, two objects, one match via name",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: dorp
  - apiVersion: v1
    kind: Pod
    metadata:
      name: foof
      namespace: derpper
  - kind: PatchTemplate
    template: |
      metadata:
        namespace: {{.Name}}
    selector:
      kinds:
      - Pod
      names:
      - foof
`,
			expMatchSubstrs:   []string{"namespace: zed", "namespace: dorp"},
			expNoMatchSubstrs: []string{"namespace: derper"},
		},
		{
			desc: "success: two patches, one object",
			opts: map[string]interface{}{
				"Name":  "zed",
				"Annot": "bar",
			},
			component: `
kind: Component
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
`,
			expMatchSubstrs:   []string{"namespace: zed", "fooAnnot: bar"},
			expNoMatchSubstrs: []string{"namespace: derp"},
		},
		{
			desc: "success: two patches, one object: filtered",
			opts: map[string]interface{}{
				"Name":  "zed",
				"Annot": "bar",
			},
			customFilter: &filter.Options{
				Annotations: map[string]string{
					"phase": "build",
				},
			},
			component: `
kind: Component
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
    metadata:
      annotations:
        phase: build
    template: |
      kind: Pod
      metadata:
        annotations:
          fooAnnot: {{.Annot}}
`,
			expMatchSubstrs:   []string{"namespace: derp", "fooAnnot: bar"},
			expNoMatchSubstrs: []string{"namespace: zed"},
		},
		{
			desc: "success: two patches, one object: filtered, with removal",
			opts: map[string]interface{}{
				"Name":  "zed",
				"Annot": "bar",
			},
			customFilter: &filter.Options{
				Annotations: map[string]string{
					"phase": "build",
				},
			},
			removeTemplates: true,
			component: `
kind: Component
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
    metadata:
      annotations:
        phase: build
    template: |
      kind: Pod
      metadata:
        annotations:
          fooAnnot: {{.Annot}}
`,
			expMatchSubstrs:   []string{"namespace: derp", "fooAnnot: bar", "{{.Name}}"},
			expNoMatchSubstrs: []string{"namespace: zed", "phase: build"},
		},
		{
			desc: "success: patch, basic options, rely on strategic-merge patch schema",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: derp
    spec:
      containers:
      - name: kube-apiserver
        image: gcr.io/google_containers/kube-apiserver:v1.9.7
      - name: kube-derp
        image: gcr.io/google_containers/derp:v1.9.7
  - kind: PatchTemplate
    template: |
      kind: Pod
      spec:
        containers:
        - name: kube-apiserver
          image: gcr.io/my/favorite:v1
`,
			expMatchSubstrs:   []string{"image: gcr.io/my/favorite:v1"},
			expNoMatchSubstrs: []string{"image: gcr.io/google_containers/kube-apiserver:v1.9.7"},
		},
		{
			desc: "success: patch, basic options, delete via strategic-merge-patch",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      namespace: derp
    spec:
      containers:
      - name: kube-apiserver
        image: gcr.io/google_containers/kube-apiserver:v1.9.7
      - name: kube-derp
        image: gcr.io/google_containers/derp:v1.9.7
  - kind: PatchTemplate
    template: |
      kind: Pod
      spec:
      containers:
      - name: kube-derp
        $patch: delete
`,
			expNoMatchSubstrs: []string{"image: gcr.io/google_containers/derp:v1.9.7"},
		},
		{
			desc: "success: patch with StrategicMergePatch patch type",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    patchType: StrategicMergePatch
    template: |
      metadata:
        namespace: foo`,
			expMatchSubstrs: []string{"metadata:\n      namespace: foo"},
			removeTemplates: true,
		},
		{
			desc: "success: unknown type, patch with JSONPatch patch type",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: TestCustomResource
  - kind: PatchTemplate
    patchType: JSONPatch
    template: |
      metadata:
        namespace: foo`,
			expMatchSubstrs: []string{
				"apiVersion: v1\n    kind: TestCustomResource\n    metadata:\n      namespace: foo",
			},
			removeTemplates: true,
		},
		{
			desc: "success: unknown type, patch with JSONPatch patch type, with options",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: TestCustomResource
  - kind: PatchTemplate
    patchType: JSONPatch
    template: |
      metadata:
        namespace: {{.Name}}`,
			expMatchSubstrs: []string{
				"apiVersion: v1\n    kind: TestCustomResource\n    metadata:\n      namespace: zed",
			},
			removeTemplates: true,
		},

		{
			desc: "success: patch with numbers (floats)",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    optionsSchema:
      properties:
        Name:
          type: string
          default: zed
        Port:
          description: Container port for the helloweb app.
          type: number
          default: 8080
    template: |
      kind: Pod
      metadata:
        namespace: {{.Name}}
      spec:
        containers:
          - name: hello-app
            {{if ge .Port 1.0}}
            ports:
            - containerPort: {{.Port}}
            {{end}}
`,
			expMatchSubstrs: []string{"containerPort: 8080"},
		},
		{
			desc: "success: patch with numbers (int to floats)",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    optionsSchema:
      properties:
        Name:
          type: string
          default: zed
        Port:
          description: Container port for the helloweb app.
          type: number
          default: 8080
    template: |
      kind: Pod
      metadata:
        namespace: {{.Name}}
      spec:
        containers:
          - name: hello-app
            {{with $x := convertAnyToFloat .Port}}{{if ge $x 1.0}}
            ports:
            - containerPort: {{$x}}
            {{end}}{{end}}
`,
			opts: map[string]interface{}{
				"Port": 8080,
			},
			expMatchSubstrs: []string{"containerPort: 8080"},
		},

		// Errors
		{
			desc: "fail: patch, basic options: missing variable",
			opts: map[string]interface{}{
				"Name": "zed",
			},
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    template: |
      kind: Pod
      metadata:
        namespace: {{.Foo}}
`,
			expErrSubstr: "map has no entry for key \"Foo\"",
		},
		{
			desc: "fail: patch, no options: (still missing variable)",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplate
    template: |
      kind: Pod
      metadata:
        namespace: {{.Foo}}
`,
			expErrSubstr: "map has no entry for key \"Foo\"",
		},
		{
			desc: "failure: unknown type, patch without JSONPatch patch type",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: TestCustomResource
  - kind: PatchTemplate
    template: |
      metadata:
        namespace: foo`,
			expErrSubstr: "kind \"TestCustomResource\" and apiVersion \"v1\": type not registered in scheme",
		},
		{
			desc: "failure: unknown type, patch with bad patch type",
			component: `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: TestCustomResource
  - kind: PatchTemplate
    patchType: PatchyMcPatchface
    template: |
      metadata:
        namespace: foo`,
			expErrSubstr: "bad patch type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			patcher := NewApplierWithConfig(WithFilterOpts(tc.customFilter), WithIncludeTemplates(!tc.removeTemplates))

			compObj, err := converter.FromYAMLString(tc.component).ToComponent()
			if err != nil {
				t.Fatalf("Error converting component %s: %v", tc.component, err)
			}

			hasErr := false
			newComp, err := patcher.ApplyOptions(compObj, tc.opts)
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

			compStr := string(compBytes)
			for _, s := range tc.expMatchSubstrs {
				if !strings.Contains(compStr, s) {
					t.Errorf("expected output yaml to contain %s", s)
					hasErr = true
				}
			}
			for _, s := range tc.expNoMatchSubstrs {
				if strings.Contains(compStr, s) {
					t.Errorf("expected output yaml to *not* contain %s", s)
					hasErr = true
				}
			}
			if hasErr {
				t.Errorf("got yaml contents:\n%s", compStr)
			}
		})
	}
}

func TestConvertToFloat(t *testing.T) {
	testCases := []struct {
		desc   string
		in     interface{}
		exp    float64
		expErr error
	}{
		{
			desc: "float success",
			in:   float64(12.2),
			exp:  12.2,
		},
		{
			desc: "float32 success",
			in:   float32(12),
			exp:  12,
		},
		{
			desc: "string success",
			in:   string("12.2"),
			exp:  12.2,
		},
		{
			desc: "int success",
			in:   int(12),
			exp:  12.0,
		},
		{
			desc: "int32 success",
			in:   int32(12),
			exp:  12.0,
		},
		{
			desc: "int64 success",
			in:   int64(12),
			exp:  12.0,
		},
		{
			desc:   "basic error",
			in:     func() {},
			expErr: floatConversionError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			out, err := convertToFloat(tc.in)
			if !errors.Is(err, tc.expErr) {
				t.Fatalf("got error %v, but expected error %v", err, tc.expErr)
			}
			if out != tc.exp {
				t.Fatalf("got value %v, but expected value %v", out, tc.exp)
			}
		})
	}
}
