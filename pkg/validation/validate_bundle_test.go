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
			bundle: `apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.9.7.testbundle'`,
			// no errors
		},

		{
			desc: "fail: duplicate image config key",
			bundle: `apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.9.7.testbundle'
  imageConfigs:
  - name: masterNode
  - name: masterNode`,
			errSubstring: "duplicate image config",
		},

		{
			desc: "fail: duplicate cluster app config key",
			bundle: `apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.9.7.testbundle'
  clusterApps:
  - name: coolApp
  - name: coolApp`,
			errSubstring: "duplicate cluster application key",
		},

		{
			desc: "fail: duplicated cluster app object config key",
			bundle: `apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.9.7.testbundle'
  clusterApps:
  - name: coolApp1
    clusterObjects:
    - name: pod
    - name: pod`,
			errSubstring: "duplicate cluster application object key",
		},

		{
			desc: "fail: no options custom resource",
			bundle: `apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '1.9.7.testbundle'
  optionsDefaults:
  - appName: foo
    objectName: bar
  clusterApps:
  - name: coolApp1
    clusterObjects:
    - name: pod`,
			errSubstring: "options specified with application",
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
