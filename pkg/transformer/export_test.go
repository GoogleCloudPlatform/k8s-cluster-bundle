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

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

const inlinedBundle = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: inlined-bundle
spec:
  clusterApps:
  - name: kubedns
    clusterObjects:
    - name: kubedns-service
      inlined:
        foo: bar
    - name: kubedns-service-account
      inlined:
        biff: bam
  - name: two-layer-app
    clusterObjects:
    - name: dynamic-control-plane-pod
      inlined:
        foo: bar
    - name: user-space-pod-1
      inlined:
        biff: bam
    - name: user-space-pod-2
      inlined:
        bar: baz
`

const filesBundle = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: files-bundle
spec:
  clusterApps:
  - name: kube-apiserver
    clusterObjects:
    - name: kube-apiserver-pod
      file:
        path: 'path/to/kube_apiserver.yaml'
`

func TestExport(t *testing.T) {
	var testcases = []struct {
		testName   string
		bundleYaml string
		appName    string
		// Export returns a list of ExportedApps that have a name and a ClusterApplication.
		// This is for checking that:
		// 1. We get applications for the layers we expect.
		// 2. The returned applications contain the objects we expect.
		expectedObjects   []string
		expectErrContains string
	}{
		{
			testName:        "single layer app",
			bundleYaml:      inlinedBundle,
			appName:         "kubedns",
			expectedObjects: []string{"kubedns-service", "kubedns-service-account"},
		},
		{
			testName:        "two layer app",
			bundleYaml:      inlinedBundle,
			appName:         "two-layer-app",
			expectedObjects: []string{"dynamic-control-plane-pod", "user-space-pod-1", "user-space-pod-2"},
		},
		{
			testName:          "cluster application not found",
			bundleYaml:        inlinedBundle,
			appName:           "not-an-app",
			expectErrContains: "not-an-app",
		},
		{
			testName:          "bundle not inlined",
			bundleYaml:        filesBundle,
			appName:           "kube-apiserver",
			expectErrContains: "not inlined",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			b, err := converter.Bundle.YAMLToProto([]byte(tc.bundleYaml))
			if err != nil {
				t.Fatalf("YAMLToProto(...) returned error: %v", err)
			}
			bp := converter.ToBundle(b)
			exporter, err := NewAppExporter(bp)
			if err != nil {
				t.Fatalf("Error creating exporter for bundle %v: %v", bp, err)
			}

			app, err := exporter.Export(bp, tc.appName)
			if tc.expectErrContains != "" {
				if err == nil {
					t.Fatalf("Export(%v, %q) should have returned an error but error was nil", bp, tc.appName)
				}
				if !strings.Contains(err.Error(), tc.expectErrContains) {
					t.Fatalf("Export(%v, %q) error message should have contained: %v, Got: %v", bp, tc.appName, tc.expectErrContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Export(%v, %v) returned unexpected error: %v", bp, tc.appName, err)
			}

			gotObjs := objectNames(app.Objects)
			if len(gotObjs) != len(tc.expectedObjects) {
				t.Errorf("Export(%v, %q) did not return the expected app, Got: %v, Want: %v", bp, tc.appName, gotObjs, tc.expectedObjects)
			}

			sort.Strings(gotObjs)
			if !reflect.DeepEqual(gotObjs, tc.expectedObjects) {
				t.Errorf("Export(%v, %q) did not return the expected objects for app %q, Got: %v, Want: %v", bp, tc.appName, app.Name, gotObjs, tc.expectedObjects)
			}
		})
	}
}

func objectNames(obj []*bpb.ClusterObject) []string {
	var out []string
	for _, o := range obj {
		out = append(out, o.GetName())
	}
	return out
}
