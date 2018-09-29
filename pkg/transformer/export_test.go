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

package transformer

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

const inlinedBundle = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: inlined-bundle
spec:
  components:
  - metadata:
      name: kubedns
    clusterObjects:
    - metadata:
        name: kubedns-service
    - metadata:
        name: kubedns-service-account

  - metadata:
      name: two-layer-app
    clusterObjects:
    - metadata:
        name: dynamic-control-plane-pod
    - metadata:
        name: user-space-pod-1
    - metadata:
        name: user-space-pod-2
`

const filesBundle = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: files-bundle
spec:
  components:
  - metadata:
      name: kube-apiserver
    clusterObjectFiles:
    - url: 'file://path/to/kube_apiserver.yaml'
`

func TestExport(t *testing.T) {
	var testcases = []struct {
		testName   string
		bundleYaml string
		compName   string
		// Export returns a list of ExportedApps that have a name and a ClusterComponent.
		// This is for checking that:
		// 1. We get components for the layers we expect.
		// 2. The returned componentscontain the objects we expect.
		expectedObjects   []string
		expectErrContains string
	}{
		{
			testName:        "single layer app",
			bundleYaml:      inlinedBundle,
			compName:        "kubedns",
			expectedObjects: []string{"kubedns-service", "kubedns-service-account"},
		},
		{
			testName:        "two layer app",
			bundleYaml:      inlinedBundle,
			compName:        "two-layer-app",
			expectedObjects: []string{"dynamic-control-plane-pod", "user-space-pod-1", "user-space-pod-2"},
		},
		{
			testName:          "cluster component not found",
			bundleYaml:        inlinedBundle,
			compName:          "not-an-app",
			expectErrContains: "not-an-app",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			b, err := converter.Bundle.YAMLToProto([]byte(tc.bundleYaml))
			if err != nil {
				t.Fatalf("YAMLToProto(...) returned error: %v", err)
			}
			bp := converter.ToBundle(b)
			exporter, err := NewComponentExporter(bp)
			if err != nil {
				t.Fatalf("Error creating exporter for bundle %v: %v", bp, err)
			}

			comp, err := exporter.Export(tc.compName)
			if tc.expectErrContains != "" {
				if err == nil {
					t.Fatalf("Export(%v, %q) should have returned an error but error was nil", bp, tc.compName)
				}
				if !strings.Contains(err.Error(), tc.expectErrContains) {
					t.Fatalf("Export(%v, %q) error message should have contained: %v, Got: %v", bp, tc.compName, tc.expectErrContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Export(%v, %v) returned unexpected error: %v", bp, tc.compName, err)
			}

			gotObjs := objectNames(comp.ClusterObjects)
			if len(gotObjs) != len(tc.expectedObjects) {
				t.Errorf("Export(%v, %q) did not return the expected component, Got: %v, Want: %v", bp, tc.compName, gotObjs, tc.expectedObjects)
			}

			sort.Strings(gotObjs)
			if !reflect.DeepEqual(gotObjs, tc.expectedObjects) {
				t.Errorf("Export(%v, %q) did not return the expected objects for component %q, Got: %v, Want: %v",
					bp, tc.compName, comp.GetMetadata().GetName(), gotObjs, tc.expectedObjects)
			}
		})
	}
}

func objectNames(obj []*structpb.Struct) []string {
	var out []string
	for _, o := range obj {
		out = append(out, core.ObjectName(o))
	}
	return out
}
