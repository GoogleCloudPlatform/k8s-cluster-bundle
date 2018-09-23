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

package cmdlib

import (
	"strings"
	"testing"
)

const bundleYAML = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: test-bundle
spec:
  nodeConfigs:
  - name: masterNode
    initFile: "echo 'I'm a script'"
    osImage:
      url: 'gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1'
  components:
  - name: kube-apiserver
    clusterObjects:
    - name: kube-apiserver-pod
      file:
        url: 'file://path/to/kube_apiserver.yaml'
`

const podYAML = `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: kube-system
spec:
  containers:
  - name: test-pod
`

const clusterBundleKind = "ClusterBundle"

func TestParseBundleYAML(t *testing.T) {
	bundleKey := "bundle"
	gibberishKey := "gibberish"
	podKey := "pod"

	yamlMap := map[string]string{
		bundleKey:    bundleYAML,
		gibberishKey: "blah " + bundleYAML + " blah",
		podKey:       podYAML,
	}

	var testcases = []struct {
		testName     string
		contentsKey  string
		expectedKind string
		expectErr    bool
	}{
		{
			testName:     "valid bundle yaml",
			contentsKey:  bundleKey,
			expectedKind: clusterBundleKind,
		},
		{
			testName:    "gibberish in bundle yaml",
			contentsKey: gibberishKey,
			expectErr:   true,
		},
		{
			testName:    "not bundle yaml",
			contentsKey: podKey,
			expectErr:   true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			reader := strings.NewReader(yamlMap[tc.contentsKey])
			b, err := parseBundleYAML(reader)
			if (tc.expectErr && err == nil) || (!tc.expectErr && err != nil) {
				t.Fatalf("parseBundleYAML(yamlMap[%q]) returned err: %v, Want Err: %v", tc.contentsKey, err, tc.expectErr)
			}
			if err != nil {
				return
			}
			if b == nil {
				t.Fatalf("No bundle returned in success case.")
			}
			if b.GetKind() != "ClusterBundle" {
				t.Errorf("parseBundleYAML(yamlMap[%q]) returned kind: %q, Want: %q", tc.contentsKey, b.GetKind(), tc.expectedKind)
			}
		})
	}
}
