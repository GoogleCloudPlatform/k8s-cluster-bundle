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
  version: '1.9.7.testbundle'
  imageConfigs:
  - name: masterNode
    initScript: "echo 'I'm a script'"
  - name: userNode
    initScript: "echo 'I'm another script'"
  clusterApps:
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
		imageName  string
		appName    string
		objName    string
		shouldFind bool
	}{
		{
			desc:       "success: bootstrap lookup",
			imageName:  "masterNode",
			shouldFind: true,
		},
		{
			desc:       "failure: bootstrap lookup",
			imageName:  "masterNoob",
			shouldFind: false,
		},
		{
			desc:       "success: cluster app lookup",
			appName:    "etcd-server",
			shouldFind: true,
		},
		{
			desc:       "failure: cluster app lookup",
			appName:    "etcd-server-bloop",
			shouldFind: false,
		},
		{
			desc:       "success: cluster obj lookup",
			appName:    "etcd-server",
			objName:    "pod",
			shouldFind: true,
		},
		{
			desc:       "failure: cluster obj lookup",
			appName:    "etcd-server",
			objName:    "blorp",
			shouldFind: false,
		},
	}
	for _, tc := range testCases {
		if tc.imageName != "" {
			v := finder.ImageConfig(tc.imageName)
			if v == nil && tc.shouldFind {
				t.Errorf("Test %v: Got unexpected nil response for lookup of bootstrap", tc.desc)
			} else if v != nil && !tc.shouldFind {
				t.Errorf("Test %v: Got unexpected non-nil response %v for lookup of bootstrap", tc.desc, v)
			}

		} else if tc.objName != "" && tc.appName != "" {
			v := finder.ClusterAppObject(tc.appName, tc.objName)
			if v == nil && tc.shouldFind {
				t.Errorf("Test %v: Got unexpected nil response for cluster app object lookup", tc.desc)
			} else if v != nil && !tc.shouldFind {
				t.Errorf("Test %v: Got unexpected non-nil response %v for cluster app object lookup", tc.desc, v)
			}

		} else if tc.appName != "" {
			v := finder.ClusterApp(tc.appName)
			if v == nil && tc.shouldFind {
				t.Errorf("Test %v: Got unexpected nil response for cluster app lookup", tc.desc)
			} else if v != nil && !tc.shouldFind {
				t.Errorf("Test %v: Got unexpected non-nil response %v for cluster app lookup", tc.desc, v)
			}

		} else {
			t.Errorf("Unexpected fallthrough for testcase %v", tc)
		}
	}
}
