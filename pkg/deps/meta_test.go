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

package deps

import (
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

func TestConvertMeta(t *testing.T) {
	testCases := []struct {
		desc      string
		component string

		expName       string
		expVersion    string
		expVisibility []string
		expMatchMeta  MatchMetadata
		expErrSubstr  string
		expRequire    bool
	}{
		{
			desc: "basic parse",
			component: `
kind: Component
spec:
  componentName: foo
  version: 0.2.0
`,
			expName:    "foo",
			expVersion: "0.2.0",
		},
		{
			desc: "requirements",
			component: `
kind: Component
spec:
  componentName: foo
  version: 0.2.0
  objects:
  - apiVersion: bundle.gke.io/v1alpha1
    kind: Requirements
    require:
    - componentName: foo
    - componentName: bar
      version: 0.2.6
`,
			expName:    "foo",
			expVersion: "0.2.0",
			expRequire: true,
		},
		{
			desc: "requirements + visibility",
			component: `
kind: Component
spec:
  componentName: foo
  version: 0.2.0
  objects:
  - apiVersion: bundle.gke.io/v1alpha1
    kind: Requirements
    require:
    - componentName: foo
    - componentName: bar
      version: 0.2.6
    visibility:
    - biff
    - bam
`,
			expName:       "foo",
			expVersion:    "0.2.0",
			expRequire:    true,
			expVisibility: []string{"biff", "bam"},
		},
		{
			desc: "annotations",
			component: `
kind: Component
metadata:
  annotations:
    foo: bar
    biff: bazz
spec:
  componentName: foo
  version: 0.2.0
`,
			expName:    "foo",
			expVersion: "0.2.0",
			expMatchMeta: &AnnotationMetadata{
				Annotations: map[string]string{
					"foo":  "bar",
					"biff": "bazz",
				},
			},
		},

		// errors
		{
			desc: "no componentName",
			component: `
kind: Component
spec:
  version: 0.2.0
`,
			expErrSubstr: "both componentName and version",
		},
		{
			desc: "no version",
			component: `
kind: Component
spec:
  componentName: foo
`,
			expErrSubstr: "both componentName and version",
		},
		{
			desc: "bad version",
			component: `
kind: Component
spec:
  componentName: foo
  version: 0.2.0.4
`,
			expErrSubstr: "while parsing version",
		},
		{
			desc: "duplicate requirements",
			component: `
kind: Component
spec:
  componentName: foo
  version: 0.2.0
  objects:
  - apiVersion: bundle.gke.io/v1alpha1
    kind: Requirements
    require:
    - componentName: foo
    - componentName: bar
      version: 0.2.6
  - apiVersion: bundle.gke.io/v1alpha1
    kind: Requirements
    require:
    - componentName: biff
    - componentName: bam
      version: 0.2.6
`,
			expErrSubstr: "duplicate requirements object",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			comp, err := converter.FromYAMLString(tc.component).ToComponent()
			if err != nil {
				t.Fatal(err)
			}

			m, err := metaFromComponent(comp, AnnotationProcessor)
			expErr := testutil.CheckErrorCases(err, tc.expErrSubstr)
			if expErr != nil {
				t.Fatal(expErr)
			}
			if err != nil {
				return
			}

			if gotName := m.componentName; tc.expName != gotName {
				t.Errorf("got componentName %q but wanted %q", gotName, tc.expVersion)
			}
			if gotVer := m.versionStr(); tc.expVersion != gotVer {
				t.Errorf("got version %q but wanted %q", gotVer, tc.expVersion)
			}
			if tc.expRequire && (m.reqDeps == nil || len(m.reqDeps) == 0) {
				t.Error("got no requirements object but expected one")
			} else if !tc.expRequire && (m.reqDeps != nil || len(m.reqDeps) > 0) {
				t.Errorf("got requirements object %+v but did not expected one", m.reqDeps)
			}
			expVisMap := make(map[string]bool)
			for _, st := range tc.expVisibility {
				expVisMap[st] = true
			}
			if !reflect.DeepEqual(expVisMap, m.visibility) {
				t.Errorf("got visibility %v but expected %v", m.visibility, expVisMap)
			}
			if tc.expMatchMeta != nil && !reflect.DeepEqual(tc.expMatchMeta, m.matchMeta) {
				t.Errorf("got matchMeta %v but expected %v", m.matchMeta, tc.expMatchMeta)
			}
		})
	}
}
