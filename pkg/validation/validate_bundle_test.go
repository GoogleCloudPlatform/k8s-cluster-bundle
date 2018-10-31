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
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: ClusterBundle
spec:
  name: '1.9.7.testbundle-zork'`,
			// no errors
		},

		{
			desc: "fail: bad kind",
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: Bundle
spec:
  name: '1.9.7.testbundle-zork'`,
			errSubstring: "bundle kind",
		},
		{
			desc: "fail: apiVersion",
			bundle: `apiVersion: 'gke.io/k8s-cluster'
kind: Bundle
spec:
  name: '1.9.7.testbundle-zork'`,
			errSubstring: "bundle apiVersion",
		},
		{
			desc: "fail: name",
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: Bundle`,
			errSubstring: "bundle name",
		},

		{
			desc: "success cluster component",
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: ClusterBundle
spec:
  name: '1.9.7.testbundle-zork'
  components:
  - apiVersion: bundle.gke.io/v1alpha1
    kind: ComponentPackage
    spec:
      name: coolApp`,
			// no errors
		},
		{
			desc: "fail cluster component: no kind",
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - apiVersion: bundle.gke.io/v1alpha1
    spec:
      name: coolApp`,
			errSubstring: "cluster component kind",
		},
		{
			desc: "fail: duplicate cluster component key",
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - spec:
      name: coolApp
  - spec:
      name: coolApp`,
			errSubstring: "duplicate cluster component key",
		},

		{
			desc: "fail: api version on cluster obj",
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - spec:
      name: coolApp1
      clusterObjects:
      - metadata:
          name: pod
        kind: zed`,
			errSubstring: "must always have an API Version",
		},

		{
			desc: "fail: no kind on cluster obj",
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - spec:
      name: coolApp1
      clusterObjects:
      - metadata:
          name: pod
        apiVersion: zed`,
			errSubstring: "must always have a kind",
		},

		{
			desc: "fail: duplicate object ref",
			bundle: `apiVersion: 'bundle.gke.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - spec:
      name: coolApp1
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
