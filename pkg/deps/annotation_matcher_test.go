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
)

var annotProcessingComps = []string{
	`
kind: Component
metadata:
spec:
  componentName: ann-none
  version: 1.0.0
`,
	`
kind: Component
metadata:
  annotations:
    cool-component: true
    qualified: true
    channel: stable
spec:
  componentName: ann
  version: 1.0.0
`,
	`
kind: Component
metadata:
  annotations:
    cool-component: true
    qualified: true
    bad-component: true
    feature: biff
    channel: beta
spec:
  componentName: ann
  version: 1.1.0
`,
	`
kind: Component
metadata:
  annotations:
    cool-component: true
    feature: bar
    channel: alpha
spec:
  componentName: ann
  version: 1.2.0
`,
}

func TestAnnotationProcessor(t *testing.T) {
	testCases := []struct {
		desc   string
		ref    bundle.ComponentReference
		annots map[string]string
	}{
		{
			desc: "ann-1.0.0",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			annots: map[string]string{
				"cool-component": "true",
				"qualified":      "true",
				"channel":        "stable",
			},
		}, {
			desc: "ann-none-1.0.0",
			ref: bundle.ComponentReference{
				ComponentName: "ann-none",
				Version:       "1.0.0",
			},
			annots: map[string]string{},
		},
	}
	comps := make(map[bundle.ComponentReference]*bundle.Component)
	for _, c := range annotProcessingComps {
		comp, err := converter.FromYAMLString(c).ToComponent()
		if err != nil {
			t.Fatal(err)
		}
		comps[comp.ComponentReference()] = comp
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mm, err := AnnotationProcessor(comps[tc.ref])
			if err != nil {
				t.Fatal(err)
			}
			meta, ok := mm.(*AnnotationMetadata)
			if !ok {
				t.Fatal("could not cast to AnnotationMetadata")
			}
			if !reflect.DeepEqual(tc.annots, meta.Annotations) {
				t.Errorf("processed annotations: got %v, but expected %v", meta.Annotations, tc.annots)
			}
		})
	}
}

func TestAnnotationMatcher(t *testing.T) {
	testCases := []struct {
		desc     string
		ref      bundle.ComponentReference
		criteria *AnnotationCriteria
		expMatch bool
	}{
		{
			desc: "ann-1.0.0, no criteria match",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{},
			expMatch: true,
		},
		{
			desc: "ann-1.0.0, match",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Match: map[string][]string{
					"channel": []string{"stable"},
				},
			},
			expMatch: true,
		},
		{
			desc: "ann-1.0.0, no match",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Match: map[string][]string{
					"channel": []string{"unstable"},
				},
			},
			expMatch: false,
		},
		{
			desc: "ann-1.0.0, multi match",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Match: map[string][]string{
					"channel": []string{"stable", "unstable"},
				},
			},
			expMatch: true,
		},
		{
			desc: "ann-1.0.0, both match",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Match: map[string][]string{
					"channel":   []string{"stable", "unstable"},
					"qualified": []string{"true"},
				},
			},
			expMatch: true,
		},
		{
			desc: "ann-1.0.0, both match",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Match: map[string][]string{
					"channel":   []string{"stable", "unstable"},
					"qualified": []string{"true"},
				},
			},
			expMatch: true,
		},
		{
			desc: "ann-none-1.0.0, match, no annotations",
			ref: bundle.ComponentReference{
				ComponentName: "ann-none",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{},
			expMatch: true,
		},
		{
			desc: "ann-none-1.0.0, no match, no annotations",
			ref: bundle.ComponentReference{
				ComponentName: "ann-none",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Match: map[string][]string{
					"channel": []string{"stable"},
				},
			},
			expMatch: false,
		},
		{
			desc: "ann-none-1.0.0, no exclude, no annotations",
			ref: bundle.ComponentReference{
				ComponentName: "ann-none",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Exclude: map[string][]string{
					"channel": []string{"stable"},
				},
			},
			expMatch: true,
		},
		{
			desc: "ann-1.0.0, exclude, annotations",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Exclude: map[string][]string{
					"channel": []string{"stable"},
				},
			},
			expMatch: false,
		},
		{
			desc: "ann-1.0.0, multi exclude - only one required",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Exclude: map[string][]string{
					"channel": []string{"stable"},
					"zorp":    []string{"derp"},
				},
			},
			expMatch: false,
		},
		{
			desc: "ann-1.0.0, both exclude and match: exclude wins ",
			ref: bundle.ComponentReference{
				ComponentName: "ann",
				Version:       "1.0.0",
			},
			criteria: &AnnotationCriteria{
				Match: map[string][]string{
					"channel": []string{"stable"},
				},
				Exclude: map[string][]string{
					"channel": []string{"stable"},
				},
			},
			expMatch: false,
		},
	}
	comps := make(map[bundle.ComponentReference]*bundle.Component)
	for _, c := range annotProcessingComps {
		comp, err := converter.FromYAMLString(c).ToComponent()
		if err != nil {
			t.Fatal(err)
		}
		comps[comp.ComponentReference()] = comp
	}

	metadata := make(map[bundle.ComponentReference]MatchMetadata)
	for _, c := range comps {
		mm, err := AnnotationProcessor(c)
		if err != nil {
			t.Fatal(err)
		}
		metadata[c.ComponentReference()] = mm
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mm, ok := metadata[tc.ref]
			if !ok {
				t.Fatalf("no metadata for ref %v", tc.ref)
			}
			if matched := AnnotationMatcher(tc.criteria)(tc.ref, mm); matched != tc.expMatch {
				t.Fatalf("for ref %v and criteria %v, got match %t but expected %t", tc.ref, tc.criteria, matched, tc.expMatch)
			}
		})
	}
}
