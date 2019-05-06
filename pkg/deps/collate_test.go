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
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
	"github.com/blang/semver"
)

var defaultCollateComponents = []string{
	`
kind: Component
spec:
  componentName: foo
  version: 1.0.1
`,
	`
kind: Component
spec:
  componentName: foo
  version: 1.0.2
`,
	`
kind: Component
spec:
  componentName: foo
  version: 1.0.3
`,
	`
kind: Component
spec:
  componentName: foo
  version: 1.0.4
`,
	`
kind: Component
spec:
  componentName: foo
  version: 1.3.4
`,
	`
kind: Component
spec:
  componentName: foo
  version: 2.3.4
`,
	`
kind: Component
spec:
  componentName: foo
  version: 2.4.4
`,
	`
kind: Component
spec:
  componentName: boo
  version: 1.0.2
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

var versionSuffix = []string{
	`
kind: Component
spec:
  componentName: foo
  version: 2.3.3-gke.0
`,
	`
kind: Component
spec:
  componentName: foo
  version: 2.3.4-gke.0
`,
	`
kind: Component
spec:
  componentName: foo
  version: 2.3.4-gke.1
`,
	`
kind: Component
spec:
  componentName: foo
  version: 2.3.4-gke.2
`,
	`
kind: Component
spec:
  componentName: foo
  version: 2.3.4-gke.12
`,
	`
kind: Component
spec:
  componentName: foo
  version: 2.4.4
`,
}

func TestLatest(t *testing.T) {
	comps, err := makeComps(defaultCollateComponents)
	if err != nil {
		t.Fatal(err)
	}
	sorted, _, err := sortedMapFromComponents(comps, AnnotationProcessor)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		desc          string
		componentName string
		matcher       Matcher

		expOut       string
		expErrSubstr string
	}{
		{
			desc:          "found latest",
			componentName: "foo",
			expOut:        "2.4.4",
		},
		{
			desc:          "found latest, boo",
			componentName: "boo",
			expOut:        "1.0.2",
		},

		{
			desc:          "annotations",
			componentName: "ann",
			expOut:        "1.2.0",
		},
		{
			desc:          "annotations",
			componentName: "ann",
			expOut:        "1.2.0",
		},
		{
			desc:          "annotations, match criteria all",
			componentName: "ann",
			expOut:        "1.2.0",
			matcher: AnnotationMatcher(&AnnotationCriteria{
				Match: map[string][]string{
					"cool-component": []string{"true"},
				},
			}),
		},
		{
			desc:          "annotations, match criteria subset",
			componentName: "ann",
			expOut:        "1.1.0",
			matcher: AnnotationMatcher(&AnnotationCriteria{
				Match: map[string][]string{
					"qualified": []string{"true"},
				},
			}),
		},
		{
			desc:          "annotations, match exactly one channel",
			componentName: "ann",
			expOut:        "1.0.0",
			matcher: AnnotationMatcher(&AnnotationCriteria{
				Match: map[string][]string{
					"channel": []string{"stable"},
				},
			}),
		},
		{
			desc:          "annotations, match one of two channels",
			componentName: "ann",
			expOut:        "1.1.0",
			matcher: AnnotationMatcher(&AnnotationCriteria{
				Match: map[string][]string{
					"channel": []string{"stable", "beta"},
				},
			}),
		},
		{
			desc:          "annotations, match and exclude criteria",
			componentName: "ann",
			expOut:        "1.0.0",
			matcher: AnnotationMatcher(&AnnotationCriteria{
				Match: map[string][]string{
					"qualified": []string{"true"},
				},
				Exclude: map[string][]string{
					"bad-component": []string{"true"},
				},
			}),
		},

		// errors
		{
			desc:          "annotations, exclude all",
			componentName: "ann",
			matcher: AnnotationMatcher(&AnnotationCriteria{
				Exclude: map[string][]string{
					"cool-component": []string{"true"},
				},
			}),
			expErrSubstr: "no latest version",
		},
		{
			desc:          "annotations, non-existent annotation",
			componentName: "ann",
			matcher: AnnotationMatcher(&AnnotationCriteria{
				Match: map[string][]string{
					"dorp": []string{"true"},
				},
			}),
			expErrSubstr: "no latest version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			vers, ok := sorted[tc.componentName]
			if !ok {
				t.Fatalf("no component found for %q", tc.componentName)
			}
			latest, err := vers.latest(tc.matcher)
			if cerr := testutil.CheckErrorCases(err, tc.expErrSubstr); cerr != nil {
				t.Fatal(cerr)
			}
			if err != nil {
				return
			}
			if got := latest.version.String(); got != tc.expOut {
				t.Errorf("for component %q, got latest version %q but expected %q", tc.componentName, got, tc.expOut)
			}
		})
	}
}

func TestPrevious(t *testing.T) {
	testCases := []struct {
		desc          string
		comps         []string
		componentName string
		version       string
		matcher       Matcher

		expOut       string
		expErrSubstr string
	}{
		{
			desc:          "strictly greater",
			componentName: "foo",
			version:       "3.0.0",
			expOut:        "2.4.4",
		},
		{
			desc:          "equiv version",
			componentName: "foo",
			version:       "2.4.4",
			expOut:        "2.3.4",
		},
		{
			desc:          "middle version",
			componentName: "foo",
			version:       "1.0.3",
			expOut:        "1.0.2",
		},
		{
			desc:          "bottom version",
			componentName: "foo",
			version:       "1.0.2",
			expOut:        "1.0.1",
		},
		{
			desc:          "single version, previous",
			componentName: "boo",
			version:       "3.0.2",
			expOut:        "1.0.2",
		},
		{
			desc:          "version suffix",
			comps:         versionSuffix,
			componentName: "foo",
			version:       "2.4.4",
			expOut:        "2.3.4-gke.12",
		},
		{
			desc:          "version suffix, previous suffix",
			comps:         versionSuffix,
			componentName: "foo",
			version:       "2.3.4-gke.12",
			expOut:        "2.3.4-gke.2",
		},
		{
			desc:          "version suffix, previous suffix, v2",
			comps:         versionSuffix,
			componentName: "foo",
			version:       "2.3.4-gke.1",
			expOut:        "2.3.4-gke.0",
		},
		{
			desc:          "version suffix, previous suffix, v2",
			comps:         versionSuffix,
			componentName: "foo",
			version:       "2.3.4-gke.0",
			expOut:        "2.3.3-gke.0",
		},

		// error cases
		{
			desc:          "error: no bottom version",
			componentName: "foo",
			version:       "1.0.1",
			expErrSubstr:  "no previous version",
		},
		{
			desc:          "error: non-existent version",
			componentName: "foo",
			version:       "0.2.1",
			expErrSubstr:  "no previous version",
		},
		{
			desc:          "error: single version, equal",
			componentName: "boo",
			version:       "1.0.2",
			expErrSubstr:  "no previous version",
		},
		{
			desc:          "error: single version, before",
			componentName: "boo",
			version:       "0.0.2",
			expErrSubstr:  "no previous version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			raw := tc.comps
			if len(raw) == 0 {
				raw = defaultCollateComponents
			}

			comps, err := makeComps(raw)
			if err != nil {
				t.Fatal(err)
			}
			sorted, _, err := sortedMapFromComponents(comps, AnnotationProcessor)
			if err != nil {
				t.Fatal(err)
			}
			vers, ok := sorted[tc.componentName]
			if !ok {
				t.Fatalf("no component found for %q", tc.componentName)
			}
			ver, err := semver.Parse(tc.version)
			if err != nil {
				t.Fatal(err)
			}

			previous, err := vers.previous(ver, tc.matcher)
			cerr := testutil.CheckErrorCases(err, tc.expErrSubstr)
			if cerr != nil {
				t.Fatal(cerr)
			}
			if err != nil {
				return
			}
			if got := previous.version.String(); got != tc.expOut {
				t.Errorf("for component %q and version %q, got previous version %q but expected %q", tc.version, tc.componentName, got, tc.expOut)
			}
		})
	}
}

func makeComps(cs []string) ([]*bundle.Component, error) {
	var out []*bundle.Component
	for _, c := range cs {
		c, err := converter.FromYAMLString(c).ToComponent()
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
