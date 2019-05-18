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

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

var twoLayerCyclic = []string{
	`
kind: Component
spec:
  componentName: foo
  version: 0.3.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: kubernetes
      version: 1.1.0`,
	`
kind: Component
spec:
  componentName: foo
  version: 0.4.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: kubernetes
      version: 1.3.0`,
	`
kind: Component
spec:
  componentName: kubernetes
  version: 1.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: foo`,
}

var threeLayerCyclic = []string{
	`
kind: Component
spec:
  componentName: low
  version: 2.3.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: high
      version: 1.3.0`,
	`
kind: Component
spec:
  componentName: low
  version: 2.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: high
      version: 1.2.0`,
	`
kind: Component
spec:
  componentName: mid
  version: 0.3.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: high
      version: 1.1.0
    - componentName: low`,
	`
kind: Component
spec:
  componentName: mid
  version: 0.4.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: high
      version: 1.3.0
    - componentName: low`,
	`
kind: Component
spec:
  componentName: high
  version: 1.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: mid`,
}

var diagonalDeps = []string{
	`
kind: Component
spec:
  componentName: low
  version: 2.3.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: high
      version: 1.2.0`,
	`
kind: Component
spec:
  componentName: low
  version: 2.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'`,
	`
kind: Component
spec:
  componentName: mid2
  version: 2.3.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: low
      version: 2.3.0`,
	`
kind: Component
spec:
  componentName: mid1
  version: 1.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: low
      version: 2.2.0`,
	`
kind: Component
spec:
  componentName: mid1
  version: 1.3.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: low
      version: 2.2.0`,
	`
kind: Component
spec:
  componentName: high
  version: 1.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: mid1
    - componentName: mid2`,
	`
kind: Component
spec:
  componentName: high
  version: 1.1.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: mid1
    - componentName: mid2`,
}

var apiVisibilityPattern = []string{
	`
kind: Component
spec:
  componentName: foo
  version: 0.3.0
  objects:
  - kind: Requirements
    require:
    - componentName: foo-api
    - componentName: kubernetes
      version: 1.2
    visibility:
    - foo-api`,
	`
kind: Component
spec:
  componentName: bad-vis
  version: 3.3.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: foo`,
	`
kind: Component
spec:
  componentName: foo-api
  version: 1.1.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: foo
      version: 0.3`,
	`
kind: Component
spec:
  componentName: kubernetes
  version: 1.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'`,
}

var annotationSet = []string{
	`
kind: Component
metadata:
  annotations:
    cool-component: true
    qualified: true
spec:
  componentName: kubernetes
  version: 1.11.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
    require:
    - componentName: ann`,
	`
kind: Component
metadata:
  annotations:
    cool-component: true
    qualified: true
spec:
  componentName: ann
  version: 1.0.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
`,
	`
kind: Component
metadata:
  annotations:
    cool-component: true
    qualified: true
    bad-component: true
spec:
  componentName: ann
  version: 1.1.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
`,
	`
kind: Component
metadata:
  annotations:
    cool-component: true
spec:
  componentName: ann
  version: 1.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
`,
	`
kind: Component
metadata:
  annotations:
    bad-component: true
spec:
  componentName: bad-component
  version: 1.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'
`,
}

func TestResolveLatest(t *testing.T) {
	testCases := []struct {
		desc     string
		universe []string
		comps    []bundle.ComponentReference
		opts     *ResolveOptions

		expComps     []bundle.ComponentReference
		expErrSubstr string
	}{
		{
			desc: "one component: exact",
			universe: []string{
				`
kind: Component
spec:
  componentName: foo
  version: 0.2.0`,
			},
			comps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.2.0"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.2.0"},
			},
		},

		{
			desc: "one component: latest",
			universe: []string{
				`
kind: Component
spec:
  componentName: foo
  version: 0.2.0`,
			},
			comps: []bundle.ComponentReference{
				{ComponentName: "foo"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.2.0"},
			},
		},

		{
			desc: "two layer, latest",
			universe: []string{
				`
kind: Component
spec:
  componentName: foo
  version: 0.2.0
  objects:
  - kind: Requirements
    visibility:
    - '@public'`,
				`
kind: Component
spec:
  componentName: foo
  version: 0.2.1
  objects:
  - kind: Requirements
    visibility:
    - '@public'`,
				`
kind: Component
spec:
  componentName: kubernetes
  version: 1.2.0
  objects:
  - kind: Requirements
    require:
    - componentName: foo
  `,
			},
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.2.1"},
				{ComponentName: "kubernetes", Version: "1.2.0"},
			},
		},

		{
			desc: "two layer, min requirement",
			universe: []string{
				`
kind: Component
spec:
  componentName: foo
  version: 0.2.0
  objects:
  - kind: Requirements
    visibility:
    - kubernetes`,
				`
kind: Component
spec:
  componentName: foo
  version: 0.2.1
  objects:
  - kind: Requirements
    visibility:
    - kubernetes`,
				`
kind: Component
spec:
  componentName: kubernetes
  version: 1.2.0
  objects:
  - kind: Requirements
    require:
    - componentName: foo
      version: 0.2.1
  `,
			},
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.2.1"},
				{ComponentName: "kubernetes", Version: "1.2.0"},
			},
		},
		{
			desc: "two layer, cyclic, simple",
			universe: []string{`
kind: Component
spec:
  componentName: foo
  version: 0.2.1
  objects:
  - kind: Requirements
    visibility:
    - kubernetes
    require:
    - componentName: kubernetes
      version: 0.2.1`,
				`
kind: Component
spec:
  componentName: kubernetes
  version: 1.2.0
  objects:
  - kind: Requirements
    visibility:
    - foo
    require:
    - componentName: foo`},
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.2.1"},
				{ComponentName: "kubernetes", Version: "1.2.0"},
			},
		},
		{
			desc:     "two layer, cyclic, downgrade",
			universe: twoLayerCyclic,
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.3.0"},
				{ComponentName: "kubernetes", Version: "1.2.0"},
			},
		},
		{
			desc:     "two layer, cyclic, two specified at top + downgrade",
			universe: twoLayerCyclic,
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
				{ComponentName: "foo"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.3.0"},
				{ComponentName: "kubernetes", Version: "1.2.0"},
			},
		},
		{
			desc:     "two layer, reversed ordering",
			universe: twoLayerCyclic,
			comps: []bundle.ComponentReference{
				{ComponentName: "foo"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.3.0"},
				{ComponentName: "kubernetes", Version: "1.2.0"},
			},
		},
		{
			desc:     "two layer, reversed ordering",
			universe: twoLayerCyclic,
			comps: []bundle.ComponentReference{
				{ComponentName: "foo"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.3.0"},
				{ComponentName: "kubernetes", Version: "1.2.0"},
			},
		},
		{
			desc:     "three layer cyclic",
			universe: threeLayerCyclic,
			comps: []bundle.ComponentReference{
				{ComponentName: "high"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "high", Version: "1.2.0"},
				{ComponentName: "mid", Version: "0.3.0"},
				{ComponentName: "low", Version: "2.2.0"},
			},
		},
		{
			desc:     "three layer cyclic, reversed",
			universe: threeLayerCyclic,
			comps: []bundle.ComponentReference{
				{ComponentName: "low"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "high", Version: "1.2.0"},
				{ComponentName: "mid", Version: "0.3.0"},
				{ComponentName: "low", Version: "2.2.0"},
			},
		},
		{
			desc:     "api pattern",
			universe: apiVisibilityPattern,
			comps: []bundle.ComponentReference{
				{ComponentName: "foo-api"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "kubernetes", Version: "1.2.0"},
				{ComponentName: "foo", Version: "0.3.0"},
				{ComponentName: "foo-api", Version: "1.1.0"},
			},
		},
		{
			desc:     "diagonal dependencies",
			universe: diagonalDeps,
			comps: []bundle.ComponentReference{
				{ComponentName: "high"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "high", Version: "1.2.0"},
				{ComponentName: "mid1", Version: "1.3.0"},
				{ComponentName: "mid2", Version: "2.3.0"},
				{ComponentName: "low", Version: "2.3.0"},
			},
		},
		{
			desc:     "re-pick diagonal dependencies",
			universe: diagonalDeps,
			comps: []bundle.ComponentReference{
				{ComponentName: "high", Version: "1.2.0"},
				{ComponentName: "mid1", Version: "1.3.0"},
				{ComponentName: "mid2", Version: "2.3.0"},
				{ComponentName: "low", Version: "2.3.0"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "high", Version: "1.2.0"},
				{ComponentName: "mid1", Version: "1.3.0"},
				{ComponentName: "mid2", Version: "2.3.0"},
				{ComponentName: "low", Version: "2.3.0"},
			},
		},
		{
			desc:     "re-pick diagonal dependencies + manual downgrade",
			universe: diagonalDeps,
			comps: []bundle.ComponentReference{
				{ComponentName: "high", Version: "1.2.0"},
				{ComponentName: "mid1", Version: "1.2.0"},
				{ComponentName: "mid2", Version: "2.3.0"},
				{ComponentName: "low", Version: "2.3.0"},
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "high", Version: "1.2.0"},
				{ComponentName: "mid1", Version: "1.2.0"},
				{ComponentName: "mid2", Version: "2.3.0"},
				{ComponentName: "low", Version: "2.3.0"},
			},
		},
		{
			desc:     "annotations: match all",
			universe: annotationSet,
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			opts: &ResolveOptions{
				Matcher: AnnotationMatcher(&AnnotationCriteria{
					Match: map[string][]string{
						"cool-component": []string{"true"},
					},
				}),
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "kubernetes", Version: "1.11.0"},
				{ComponentName: "ann", Version: "1.2.0"},
			},
		},
		{
			desc:     "annotations: match qualified",
			universe: annotationSet,
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			opts: &ResolveOptions{
				Matcher: AnnotationMatcher(&AnnotationCriteria{
					Match: map[string][]string{
						"cool-component": []string{"true"},
						"qualified":      []string{"true"},
					},
				}),
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "kubernetes", Version: "1.11.0"},
				{ComponentName: "ann", Version: "1.1.0"},
			},
		},
		{
			desc:     "annotations: match qualified, exclude bad",
			universe: annotationSet,
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			opts: &ResolveOptions{
				Matcher: AnnotationMatcher(&AnnotationCriteria{
					Match: map[string][]string{
						"cool-component": []string{"true"},
						"qualified":      []string{"true"},
					},
					Exclude: map[string][]string{
						"bad-component": []string{"true"},
					},
				}),
			},
			expComps: []bundle.ComponentReference{
				{ComponentName: "kubernetes", Version: "1.11.0"},
				{ComponentName: "ann", Version: "1.0.0"},
			},
		},

		////////////
		// errors //
		////////////
		{
			desc: "default private prevents being depended on",
			universe: []string{
				`
kind: Component
spec:
  componentName: foo
  version: 1.2.3
`,
				`
kind: Component
spec:
  componentName: bar
  version: 2.0.0
  objects:
  - kind: Requirements
    require:
    - componentName: foo`,
			},
			comps: []bundle.ComponentReference{
				{ComponentName: "bar"},
			},
			expErrSubstr: "not visible",
		},
		{
			desc:     "incompatible component combination from requirements",
			universe: twoLayerCyclic,
			comps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.4.0"},
			},
			expErrSubstr: "component was fixed",
		},
		{
			desc: "can't downgrade",
			universe: []string{
				`
kind: Component
spec:
  componentName: foo
  version: 1.2.3
`,
				`
kind: Component
spec:
  componentName: bar
  version: 2.0.0
  objects:
  - kind: Requirements
    require:
    - componentName: foo
      version: 1.3`,
			},
			comps: []bundle.ComponentReference{
				{ComponentName: "bar"},
			},
			expErrSubstr: "no previous version to 2.0.0",
		},
		{
			desc:     "unknown component",
			universe: twoLayerCyclic,
			comps: []bundle.ComponentReference{
				{ComponentName: "foo", Version: "0.5.0"},
			},
			expErrSubstr: "unknown component",
		},
		{
			desc:     "bad visibility",
			universe: apiVisibilityPattern,
			comps: []bundle.ComponentReference{
				{ComponentName: "bad-vis"},
			},
			expErrSubstr: "not visible",
		},
		{
			desc:     "diagonal dependencies: can't downgrade",
			universe: diagonalDeps,
			comps: []bundle.ComponentReference{
				{ComponentName: "high", Version: "1.1.0"},
			},
			expErrSubstr: "found incompatibility",
		},
		{
			desc:     "annotations: can't match all",
			universe: annotationSet,
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			opts: &ResolveOptions{
				Matcher: AnnotationMatcher(&AnnotationCriteria{
					Match: map[string][]string{
						"bad-component": []string{"true"},
					},
				}),
			},
			expErrSubstr: "no latest version",
		},
		{
			desc:     "annotations: bad fixed version",
			universe: annotationSet,
			comps: []bundle.ComponentReference{
				{ComponentName: "bad-component", Version: "1.2.0"},
			},
			opts: &ResolveOptions{
				Matcher: AnnotationMatcher(&AnnotationCriteria{
					Exclude: map[string][]string{
						"bad-component": []string{"true"},
					},
				}),
			},
			expErrSubstr: "does not match the matcher conditions",
		},
		{
			desc:     "annotations: can't match any",
			universe: annotationSet,
			comps: []bundle.ComponentReference{
				{ComponentName: "kubernetes"},
			},
			opts: &ResolveOptions{
				Matcher: AnnotationMatcher(&AnnotationCriteria{
					Match: map[string][]string{
						"zorp": []string{"true"},
					},
				}),
			},
			expErrSubstr: "no latest version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			var parsedComps []*bundle.Component
			for _, str := range tc.universe {
				comp, err := converter.FromYAMLString(str).ToComponent()
				if err != nil {
					t.Fatal(err)
				}
				parsedComps = append(parsedComps, comp)
			}

			resolver, err := NewResolver(parsedComps, AnnotationProcessor)
			if err != nil {
				t.Fatal(err)
			}

			expMap := make(map[string]string)
			for _, ec := range tc.expComps {
				expMap[ec.ComponentName] = ec.Version
			}

			got, err := resolver.Resolve(tc.comps, tc.opts)
			if cerr := testutil.CheckErrorCases(err, tc.expErrSubstr); cerr != nil {
				t.Fatal(cerr)
			}
			if err != nil {
				return
			}

			gotMap := make(map[string]string)
			for _, gc := range got {
				gotMap[gc.ComponentName] = gc.Version
			}

			if !reflect.DeepEqual(gotMap, expMap) {
				t.Errorf("Resolve(%v)=%v. Expected %v", tc.comps, got, tc.expComps)
			}
		})
	}
}
