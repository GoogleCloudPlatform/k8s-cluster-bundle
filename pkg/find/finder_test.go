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

package find

import (
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

var validBundleExample = `apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  nodeConfigs:
  - name: masterNode
    initFile: "echo 'I'm a script'"
  - name: userNode
    initFile: "echo 'I'm another script'"
  components:
  - name: etcd-server
    clusterObjects:
    - name: pod
    - name: dwerp
  - name: kube-apiserver
    clusterObjects:
    - name: pod
`

func TestBundleFinder(t *testing.T) {
	b, err := converter.Bundle.YAMLToProto([]byte(validBundleExample))
	if err != nil {
		t.Fatalf("error converting bundle: %v", err)
	}
	finder, err := NewBundleFinder(converter.ToBundle(b))
	if err != nil {
		t.Fatalf("error creating bundle finder: %v", err)
	}
	testCases := []struct {
		desc       string
		nodeName   string
		compName   string
		objName    string
		shouldFind bool
	}{
		{
			desc:       "success: bootstrap lookup",
			nodeName:   "masterNode",
			shouldFind: true,
		},
		{
			desc:       "failure: bootstrap lookup",
			nodeName:   "masterNoob",
			shouldFind: false,
		},
		{
			desc:       "success: cluster comp lookup",
			compName:   "etcd-server",
			shouldFind: true,
		},
		{
			desc:       "failure: cluster comp lookup",
			compName:   "etcd-server-bloop",
			shouldFind: false,
		},
		{
			desc:       "success: cluster obj lookup",
			compName:   "etcd-server",
			objName:    "pod",
			shouldFind: true,
		},
		{
			desc:       "failure: cluster obj lookup",
			compName:   "etcd-server",
			objName:    "blorp",
			shouldFind: false,
		},
	}
	for _, tc := range testCases {
		if tc.nodeName != "" {
			v := finder.NodeConfig(tc.nodeName)
			if v == nil && tc.shouldFind {
				t.Errorf("Test %v: Got unexpected nil response for lookup of bootstrap", tc.desc)
			} else if v != nil && !tc.shouldFind {
				t.Errorf("Test %v: Got unexpected non-nil response %v for lookup of bootstrap", tc.desc, v)
			}

		} else if tc.objName != "" && tc.compName != "" {
			v := finder.ClusterComponentObject(tc.compName, tc.objName)
			if v == nil && tc.shouldFind {
				t.Errorf("Test %v: Got unexpected nil response for cluster comp object lookup", tc.desc)
			} else if v != nil && !tc.shouldFind {
				t.Errorf("Test %v: Got unexpected non-nil response %v for cluster comp object lookup", tc.desc, v)
			}

		} else if tc.compName != "" {
			v := finder.ClusterComponent(tc.compName)
			if v == nil && tc.shouldFind {
				t.Errorf("Test %v: Got unexpected nil response for cluster comp lookup", tc.desc)
			} else if v != nil && !tc.shouldFind {
				t.Errorf("Test %v: Got unexpected non-nil response %v for cluster comp lookup", tc.desc, v)
			}

		} else {
			t.Errorf("Unexpected fallthrough for testcase %v", tc)
		}
	}
}
