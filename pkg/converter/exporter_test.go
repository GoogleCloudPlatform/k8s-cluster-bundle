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

package converter

import (
	"strings"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	objForExport = []string{`
apiVersion: v1
kind: Pod
metadata:
  name: rescheduler
`, `
apiVersion: v1
kind: Pod
metadata:
  name: etcd
`}
)

func TestExporterMulti(t *testing.T) {
	var obj []*unstructured.Unstructured
	for _, o := range objForExport {
		un, err := FromYAMLString(o).ToUnstructured()
		if err != nil {
			t.Fatal(err)
		}
		obj = append(obj, un)
	}
	comp := &bundle.Component{
		Spec: bundle.ComponentSpec{
			Objects: obj,
		},
	}
	multi, err := NewExporter(comp).ObjectsAsMultiYAML()
	if err != nil {
		t.Fatalf("Failed to multi-export yaml: %v", err)
	}
	if len(multi) != 2 {
		t.Fatalf("Got items %v, but expected exactly 2", multi)
	}
}

func TestExporterSingle(t *testing.T) {
	var obj []*unstructured.Unstructured
	for _, o := range objForExport {
		un, err := FromYAMLString(o).ToUnstructured()
		if err != nil {
			t.Fatal(err)
		}
		obj = append(obj, un)
	}
	comp := &bundle.Component{
		Spec: bundle.ComponentSpec{
			Objects: obj,
		},
	}
	single, err := NewExporter(comp).ObjectsAsSingleYAML()
	if err != nil {
		t.Fatalf("failed to single-export yaml: %v", err)
	}
	if !strings.Contains(single, "\n---\n") {
		t.Errorf("Got %s, but expected yaml to contain document join string", single)
	}
	if !strings.Contains(single, "rescheduler") {
		t.Errorf("Got %s, but expected yaml to contain 'rescheduler'", single)
	}
	if !strings.Contains(single, "etcd") {
		t.Errorf("Got %s, but expected yaml to contain 'etcd'", single)
	}
}

func TestComponentsWithPathTemplates(t *testing.T) {
	defaultComps := []string{
		`
apiVersion: bundle.gke.io/v1alpha1
kind: Component
metadata:
  name: foo-component
spec:
  componentName: foo
  version: 1.0.0`,
		`apiVersion: bundle.gke.io/v1alpha1
kind: Component
metadata:
  name: bar-component
spec:
  componentName: bar
  version: 2.0.0`,
	}
	testCases := []struct {
		desc          string
		comps         []string
		compSet       string
		pathTemplates []string

		expPathToSubstrs map[string]string
		expErrSubstr     string
	}{
		{
			desc:  "success",
			comps: defaultComps,
			pathTemplates: []string{
				"{{.ComponentName}}/{{.Version}}/component.yaml",
			},
			expPathToSubstrs: map[string]string{
				"foo/1.0.0/component.yaml": "name: foo-component",
				"bar/2.0.0/component.yaml": "name: bar-component",
			},
		},
		{
			desc:  "success: ignored paths",
			comps: defaultComps,
			pathTemplates: []string{
				"{{.ComponentName}}/{{.Version}}/{{.Blah}}/component.yaml",
				"{{.ComponentName}}/{{.Version}}/component.yaml",
			},
			expPathToSubstrs: map[string]string{
				"foo/1.0.0/component.yaml": "name: foo-component",
				"bar/2.0.0/component.yaml": "name: bar-component",
			},
		},
		{
			desc: "success: build tags",
			comps: []string{
				`
apiVersion: bundle.gke.io/v1alpha1
kind: Component
metadata:
  name: foo-component
  annotations:
    bundle.gke.io/build-tags: latest
spec:
  componentName: foo
  version: 1.0.0`,
			},
			pathTemplates: []string{
				"{{.ComponentName}}/{{.BuildTag}}/component.yaml",
				"{{.ComponentName}}/{{.Version}}/component.yaml",
			},
			expPathToSubstrs: map[string]string{
				"foo/1.0.0/component.yaml":  "name: foo-component",
				"foo/latest/component.yaml": "name: foo-component",
			},
		},
		{
			desc: "success: multiple build tags",
			comps: []string{
				`
apiVersion: bundle.gke.io/v1alpha1
kind: Component
metadata:
  name: foo-component
  annotations:
    bundle.gke.io/build-tags: latest,stable
spec:
  componentName: foo
  version: 1.0.0`,
			},
			pathTemplates: []string{
				"{{.ComponentName}}/{{.BuildTag}}/component.yaml",
				"{{.ComponentName}}/{{.BuildTag}}/component.yaml",
				"{{.ComponentName}}/{{.Version}}/component.yaml",
			},
			expPathToSubstrs: map[string]string{
				"foo/1.0.0/component.yaml":  "name: foo-component",
				"foo/latest/component.yaml": "name: foo-component",
				"foo/stable/component.yaml": "name: foo-component",
			},
		},
		{
			desc: "success: multiple build tags, component set",
			comps: []string{
				`
apiVersion: bundle.gke.io/v1alpha1
kind: Component
metadata:
  name: foo-component
  annotations:
    bundle.gke.io/build-tags: latest
spec:
  componentName: foo
  version: 1.0.0`,
			},
			compSet: `
apiVersion: bundle.gke.io/v1alpha1
kind: Component
spec:
  setName: foo-set
  version: 3.0.0
`,
			pathTemplates: []string{
				"components/{{.ComponentName}}/{{.BuildTag}}/component.yaml",
				"components/{{.ComponentName}}/{{.Version}}/component.yaml",
				"sets/{{.SetName}}/{{.Version}}/set.yaml",
			},
			expPathToSubstrs: map[string]string{
				"components/foo/1.0.0/component.yaml":  "name: foo-component",
				"components/foo/latest/component.yaml": "name: foo-component",
				"sets/foo-set/3.0.0/set.yaml":          "setName: foo-set",
			},
		},
		{
			desc: "fail: can't parse pathTemplate",
			comps: []string{
				`
apiVersion: bundle.gke.io/v1alpha1
kind: Component
metadata:
  name: foo-component
  annotations:
    bundle.gke.io/build-tags: latest
spec:
  componentName: foo
  version: 1.0.0`,
			},
			pathTemplates: []string{
				"components/{{.blaaah",
			},
			expErrSubstr: "while parsing path template",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			var comps []*bundle.Component
			for _, cstr := range tc.comps {
				comp, err := FromYAMLString(cstr).ToComponent()
				if err != nil {
					t.Fatal(err)
				}
				comps = append(comps, comp)
			}
			var cset *bundle.ComponentSet
			if tc.compSet != "" {
				cs, err := FromYAMLString(tc.compSet).ToComponentSet()
				if err != nil {
					t.Fatal(err)
				}
				cset = cs
			}
			outMap, err := NewExporter(comps...).ComponentsWithPathTemplates(tc.pathTemplates, cset)
			cerr := testutil.CheckErrorCases(err, tc.expErrSubstr)
			if cerr != nil {
				t.Fatal(cerr)
			}
			if err != nil {
				// even though cerr checks out, an err is terminal: we can't check contents.
				return
			}
			for expKey, expSubstr := range tc.expPathToSubstrs {
				val, ok := outMap[expKey]
				if !ok {
					t.Errorf("path key was empty, expected path key %q", expKey)
					continue
				}
				if !strings.Contains(val, expSubstr) {
					t.Errorf("got %q=%s, but expected value at to contain %q", expKey, val, expSubstr)
				}
			}
			for foundKey, val := range outMap {
				if _, ok := tc.expPathToSubstrs[foundKey]; !ok {
					t.Errorf("got %q=%s, but path key was not in the expected substring map", foundKey, val)
				}
			}
		})
	}
}
