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

package validate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

var defaultComponentSet = `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  setName: foo-set
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2
`

var defaultComponentSetNoRefs = `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  setName: bar-set
  version: 1.0.2
`

var defaultComponentData = `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: bar-comp-2.0.3
  spec:
    componentName: foo-comp
    version: 2.0.3
    appVersion: 2.4.5
`

func TestValidateAll(t *testing.T) {
	testCases := []struct {
		desc         string
		set          string
		components   string
		errSubstring string
	}{
		{
			desc:       "success",
			set:        defaultComponentSet,
			components: defaultComponentData,
			// no errors
		},

		// Tests for component sets
		{
			desc: "component set fail: bad kind",
			set: `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Zor
spec:
  setName: zip
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2`,
			components:   defaultComponentData,
			errSubstring: "must be ComponentSet",
		},

		{
			desc: "component set fail: apiVersion",
			set: `
apiVersion: 'zork.gke.io/v1alpha1'
kind: ComponentSet
spec:
  setName: zip
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2`,
			components:   defaultComponentData,
			errSubstring: "bundle.gke.io/<version>",
		},
		{
			desc: "component set fail: invalid X.Y.Z version string",
			set: `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  setName: zip
  version: foo
  components:
  - componentName: foo-comp
    version: 1.0.2`,
			components:   defaultComponentData,
			errSubstring: "must be of the form X.Y.Z",
		},
		{
			desc: "fail: missing X.Y.Z version string",
			set: `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  setName: zip
  components:
  - componentName: foo-comp
    version: 1.0.2`,
			components:   defaultComponentData,
			errSubstring: "Required value",
		},
		{
			desc: "fail: missing set name",
			set: `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2`,
			components:   defaultComponentData,
			errSubstring: "Required value",
		},

		// Tests for Components
		{
			desc: "success",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.10.1
    appVersion: 3.10.1`,
		},
		{
			desc: "success: X.Y app version",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.10.1
    appVersion: 3.10`,
		},
		{
			desc: "success: X.Y.Z-blah app version",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.10.1
    appVersion: 3.10.32-blah.0`,
		},
		{
			desc: "fail component: no kind",
			set:  defaultComponentSet,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2`,
			errSubstring: "must be ComponentPackage",
		},
		{
			desc: "fail component: duplicate component reference",
			set:  defaultComponentSet,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2-zed
  spec:
    componentName: foo-comp
    version: 1.0.2`,
			errSubstring: "component reference",
		},

		{
			desc: "fail: component invalid X.Y.Z version string ",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.010.1`,
			errSubstring: "must be of the form X.Y.Z",
		},
		{
			desc: "fail: component invalid X.Y.Z app version string ",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.10.1,
    appVersion: 2.010.1`,
			errSubstring: "must be of the form X.Y.Z",
		},
		{
			desc: "fail: component missing X.Y.Z version string ",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp`,
			errSubstring: "Required value",
		},

		{
			desc: "component object success",
			set:  defaultComponentSet,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: foo-pod`,
		},
		{
			desc: "object fail: duplicate",
			set:  defaultComponentSet,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: foo-pod
    - apiVersion: v1
      kind: Pod
      metadata:
        name: foo-pod`,
			errSubstring: "object reference",
		},
		{
			desc: "object fail: no metadata.name",
			set:  defaultComponentSet,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: ComponentPackage
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2
    objects:
    - apiVersion: v1
      kind: Pod`,
			errSubstring: "Required value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			set, err := converter.FromYAMLString(tc.set).ToComponentSet()
			if err != nil {
				t.Fatalf("error converting component set: %v", err)
			}
			comp, err := converter.FromYAMLString(tc.components).ToBundle()
			if err != nil {
				t.Fatalf("error converting component data: %v. was:\n%s", err, tc.components)
			}
			if err = checkErrCases(All(comp.Components, set).ToAggregate(), tc.errSubstring); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func checkErrCases(err error, expErrSubstring string) error {
	if err == nil && expErrSubstring == "" {
		return nil // success!
	} else if err == nil && expErrSubstring != "" {
		return fmt.Errorf("got nil error, but expected one containing %q", expErrSubstring)
	} else if err != nil && expErrSubstring == "" {
		return fmt.Errorf("got error: %q. but did not expect one", err.Error())
	} else if err != nil && expErrSubstring != "" && !strings.Contains(err.Error(), expErrSubstring) {
		return fmt.Errorf("got error: %q. expected it to contain substring %q", err.Error(), expErrSubstring)
	}
	return nil
}

func TestValidateAll_MultipleErrors(t *testing.T) {
	testCases := []struct {
		desc       string
		set        string
		components string
		numErrors  int
	}{
		{
			desc:       "success: no errors",
			set:        defaultComponentSet,
			components: defaultComponentData,
		},
		{
			desc: "set errors: apiversion, kind, component-mismatch",
			set: `
spec:
  setName: zip
  version: 1.0.2
  components:
  - componentName: zap-comp
    version: 1.0.2`,
			components: defaultComponentData,
			numErrors:  3,
		},
		{
			desc: "set and component errors",
			set: `
spec:
  setName: zip
  version: 1.0.2
  components:
  - componentName: zap-comp
    version: 1.0.2`,
			components: `
components:
- spec:
    componentName: foo-comp
    version: 1.0.2
    objects:
    - apiVersion: v1
      kind: Pod`,
			numErrors: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			set, err := converter.FromYAMLString(tc.set).ToComponentSet()
			if err != nil {
				t.Fatalf("error converting component set: %v", err)
			}
			comp, err := converter.FromYAMLString(tc.components).ToBundle()
			if err != nil {
				t.Fatalf("error converting component data: %v. was:\n%s", err, tc.components)
			}
			if errs := All(comp.Components, set); len(errs) != tc.numErrors {
				t.Errorf("got %d errors. expected exactly %d. errors were: %v", len(errs), tc.numErrors, errs)
			}
		})
	}
}
