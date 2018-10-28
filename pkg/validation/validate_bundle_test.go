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

package validation

import (
	"fmt"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

func TestValidateBundle(t *testing.T) {
	testCases := []struct {
		desc         string
		bundle       string
		errSubstring string
	}{
		{
			desc: "success",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.0.0'`,
			// no errors
		},

		{
			desc: "fail: bad kind",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: Bundle
metadata:
  name: '1.9.7.testbundle-zork'`,
			errSubstring: "bundle kind",
		},
		{
			desc: "fail: apiVersion",
			bundle: `apiVersion: 'gke.io/k8s-cluster'
kind: Bundle
metadata:
  name: '1.9.7.testbundle-zork'`,
			errSubstring: "bundle apiVersion",
		},
		{
			desc: "fail: name",
			bundle: `apiVersion: 'gke.io/k8s-cluster'
kind: Bundle`,
			errSubstring: "bundle name",
		},
		{
			desc: "fail: missing bundle version",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'`,
			errSubstring: "cluster bundle spec version string",
		},
		{
			desc: "fail: cluster bundle version is not SemVer",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '2'`,
			errSubstring: "cluster bundle spec version string",
		},

		{
			desc: "success cluster component",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.10.2-something+else'
  components:
  - apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
    kind: ComponentPackage
    spec:
      version: '1.1.2-staging+12345'
      requirements:
      - component: 'libCool'
        componentApiVersion: '1.8.0'
    metadata:
      name: coolApp
  - apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
    kind: ComponentPackage
    spec:
      version: '1.8.6'
    metadata:
      name: libCool`,
			// no errors
		},
		{
			desc: "fail cluster component: no kind",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
    metadata:
      name: coolApp`,
			errSubstring: "cluster component kind",
		},
		{
			desc: "fail: duplicate cluster component key",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - metadata:
      name: coolApp
  - metadata:
      name: coolApp`,
			errSubstring: "duplicate cluster component key",
		},
		{
			desc: "fail cluster component: missing SemVer version",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.10.2-something+else'
  components:
  - apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
    kind: ComponentPackage
    metadata:
      name: coolApp`,
			errSubstring: "cluster component spec version",
		},
		{
			desc: "fail cluster component: invalid SemVer string",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.10.2-something+else'
  components:
  - apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
    kind: ComponentPackage
    spec:
      version: '1.01.2'
    metadata:
      name: coolApp`,
			errSubstring: "cluster component spec version",
		},

		{
			desc: "fail: api version on cluster obj",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - metadata:
      name: coolApp1
    spec:
      clusterObjects:
      - metadata:
          name: pod
        kind: zed`,
			errSubstring: "must always have an API Version",
		},
		{
			desc: "fail: invalid SemVer string on min requirements element",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.10.2-something+else'
  components:
  - apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
    kind: ComponentPackage
    spec:
      version: '1.1.2-staging+12345'
      requirements:
      - component: 'libCool'
        componentApiVersion: '1.8.invalid'
    metadata:
      name: coolApp
  - apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
    kind: ComponentPackage
    spec:
      version: '1.8.6'
    metadata:
      name: libCool`,
			errSubstring: "min requirement has invalid SemVer string",
		},

		{
			desc: "fail: no kind on cluster obj",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - metadata:
      name: coolApp1
    spec:
      clusterObjects:
      - metadata:
          name: pod
        apiVersion: zed`,
			errSubstring: "must always have a kind",
		},

		{
			desc: "fail: duplicate object ref",
			bundle: `apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - metadata:
      name: coolApp1
    spec:
      clusterObjects:
      - metadata:
          name: pod
        apiVersion: zed,
        kind: zork
      - metadata:
          name: pod
        apiVersion: zed,
        kind: zork`,
			errSubstring: "duplicate cluster object found",
		},
	}

	for _, tc := range testCases {
		b, err := converter.Bundle.YAMLToProto([]byte(tc.bundle))
		if err != nil {
			t.Fatalf("error converting bundle: %v", err)
		}
		bval := NewBundleValidator(converter.ToBundle(b))

		if err = checkErrCases(tc.desc, err, ""); err != nil {
			t.Errorf(err.Error())
			continue
		}
		if err = checkErrCases(tc.desc, JoinErrors(bval.Validate()), tc.errSubstring); err != nil {
			t.Errorf(err.Error())
		}
	}
}

func checkErrCases(desc string, err error, expErrSubstring string) error {
	if err == nil && expErrSubstring == "" {
		return nil // success!
	} else if err == nil && expErrSubstring != "" {
		return fmt.Errorf("Test %q: Got nil error, but one containing %q", desc, expErrSubstring)
	} else if err != nil && expErrSubstring == "" {
		return fmt.Errorf("Test %q: Got error: %q. but did not expect one", desc, err.Error())
	} else if err != nil && expErrSubstring != "" && !strings.Contains(err.Error(), expErrSubstring) {
		return fmt.Errorf("Test %q: Got error: %q. expected it to contain substring %q", desc, err.Error(), expErrSubstring)
	}
	return nil
}

func TestSemVerPattern(t *testing.T) {
	testCases := []struct {
		version string
		matches bool
	}{
		{"0.0.0", true},
		{"1.0.0", true},
		{"1.2.3", true},
		{"1.10.2-alpha", true},
		{"1.10.3-alpha.beta-1.gamma", true},
		{"1.4.5+1234", true},
		{"1.2.6+12.45.alpha-1.beta2", true},
		{"1.0.1-v1alpha1+1234", true},
		{"10.22.3123-v1alpha-1.2+metadata-2.2-k32.v1", true},
		{"version", false},
		{"1", false},
		{"12.", false},
		{"3.1", false},
		{"4.32.", false},
		{"10.2.2.", false},
		{"01.2.1", false},
		{"1.02.1", false},
		{"00.0.0", false},
		{"2.0.003", false},
		{"1.2.3-alpha_1", false},
		{"1.2.3+124@23", false},
	}
	for _, tc := range testCases {
		gotMatch := semVerPattern.MatchString(tc.version)
		if gotMatch != tc.matches {
			t.Errorf("'%v', got: %v, want: %v\n", tc.version, gotMatch, tc.matches)
		}
	}
}
