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
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
)

var bundleExampleAll = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  nodeConfigs:
  - name: 'ubuntu-control-plane'
    initFile: "echo 'I'm a script'"
    osImage:
      url: 'gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1'
    envVars:
      - name: FOO_VAR
        value: 'foo-val'
  - name: 'ubuntu-cluster-node'
    initFile: "echo 'I'm another script'"
    osImage:
      url: 'gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1'
  - name: 'ubuntu-cluster-node-no-image'
    initFile: "echo 'I'm another script'"
  components:
  - metadata:
      name: logger
    clusterObjects:
    - name: logger-pod
      inline:
        apiVersion: v1
        kind: Pod
        metadata:
          name: logger-pod
        spec:
          dnsPolicy: Default
          containers:
          - name: logger
            image: gcr.io/floof/logger
            command:
               - /logger
               - --logtostderr
          - name: chopper
            image: gcr.io/floof/chopper
            command:
               - /chopper
               - --logtostderr
  - metadata:
      name: zap
    clusterObjects:
    - inline:
        apiVersion: v1
        kind: Pod
        metadata:
          name: zap
  - metadata:
      name: dap
    clusterObjects:
    - inline:
        apiVersion: v1
        kind: Pod
        metadata:
          name: dap-pod
        spec:
          containers:
          - name: dapper
            image: gcr.io/floof/dapper
          - name: verydapper
            image: gcr.io/floof/dapper`

func TestTransformStringSub(t *testing.T) {
	s, err := converter.Bundle.YAMLToProto([]byte(bundleExampleAll))
	if err != nil {
		t.Fatalf("error converting bundle: %v", err)
	}
	bun := converter.ToBundle(s)
	trans := &ImageTransformer{bun}
	rules := []*ImageSubRule{{
		Find:    "gcr.io",
		Replace: "k8s.io",
	}, {
		Find:    "gs://",
		Replace: "go://",
	}, {
		Find:    "/chopper",
		Replace: "/flopper",
	}, {
		Find:    "/dapper",
		Replace: "/mapper",
	}}

	newb := trans.TransformImagesStringSub(rules)

	found := (&find.ImageFinder{converter.ToBundle(newb)}).AllImages().Flattened()
	expected := &find.AllImagesFlattened{
		NodeImages: []string{
			"go://google-images/ubuntu/ubuntu-1604-xenial-20180509-1",
		},
		ContainerImages: []string{
			"k8s.io/floof/logger",
			"k8s.io/floof/flopper",
			"k8s.io/floof/mapper",
		},
	}
	if !reflect.DeepEqual(found, expected) {
		t.Errorf("For finding all images after string substitution, got %v, but wanted %v", found, expected)
	}

	// Make sure the old bundle didn't change
	oldFound := (&find.ImageFinder{converter.ToBundle(bun)}).AllImages().Flattened()
	oldExpected := &find.AllImagesFlattened{
		NodeImages: []string{
			"gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1",
		},
		ContainerImages: []string{
			"gcr.io/floof/logger",
			"gcr.io/floof/chopper",
			"gcr.io/floof/dapper",
		},
	}
	if !reflect.DeepEqual(oldFound, oldExpected) {
		t.Errorf("For finding all images after string substitution, the old bundle should not change. Got %v, but wanted %v", found, expected)
	}
}
