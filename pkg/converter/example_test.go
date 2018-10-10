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

package converter

import (
	"testing"
)

const bundleSimple = `
apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: test-bundle
spec:
  nodeConfigs:
  - metadata:
      name: master
    initFile: "echo 'making a control plane'"
  components:
  - metadata:
      name: kube-apiserver
    clusterObjectFiles:
    - url: 'file://path/to/kube_apiserver.yaml'
`

func TestBundleParse(t *testing.T) {
	b, err := Bundle.YAMLToProto([]byte(bundleSimple))
	if err != nil {
		t.Fatalf("Error parsing bundle: %v", err)
	}
	bp := ToBundle(b)

	if bp.GetMetadata().GetName() != "test-bundle" {
		t.Errorf("Got name %q, expected name %q", bp.Metadata.Name, "test-bundle")
	}
}
