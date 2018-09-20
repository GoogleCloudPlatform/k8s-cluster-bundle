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
	"fmt"
	"strings"
	"testing"

	"context"
	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
)

const bundleWithRefs = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: test-bundle
spec:
  nodeConfigs:
  - name: master
    externalInitFile:
      url: 'file://path/to/init-script.sh'
  components:
  - name: kube-apiserver
    clusterObjects:
    - name: pod
      file:
        url: 'file://path/to/kube_apiserver.yaml'
  - name: kubelet-config
    clusterObjects:
    - name: kubelet-config-pod
      file:
        url: 'file://path/to/kubelet/config.yaml'
`

const (
	initScriptFile    = "path/to/init-script.sh"
	kubeletConfigFile = "path/to/kubelet/config.yaml"
	kubeAPIServerFile = "path/to/kube_apiserver.yaml"

	initScriptContents = "#!/bin/bash\necho foo"
)

type fakeLocalReader struct{}

func (*fakeLocalReader) ReadFile(ctx context.Context, file *bpb.File) ([]byte, error) {
	url := file.GetUrl()
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	switch {
	case url == initScriptFile:
		return []byte(initScriptContents), nil

	case url == kubeletConfigFile:
		return []byte("{\"foo\": \"bar\"}"), nil

	case url == kubeAPIServerFile:
		return []byte("{\"biff\": \"bam\"}"), nil
	default:
		return nil, fmt.Errorf("unexpected file path %q", file.GetUrl())
	}
}

func TestInlineBundle(t *testing.T) {
	ctx := context.Background()
	b, err := converter.Bundle.YAMLToProto([]byte(bundleWithRefs))
	if err != nil {
		t.Fatalf("Error parsing bundle: %v", err)
	}
	bp := converter.ToBundle(b)
	inliner := &Inliner{
		LocalReader: &fakeLocalReader{},
	}

	newpb, err := inliner.Inline(ctx, bp)
	if err != nil {
		t.Fatalf("Error inlining bundle: %v", err)
	}
	finder, err := find.NewBundleFinder(newpb)
	if err != nil {
		t.Fatalf("Error creating bundle finder: %v", err)
	}
	if got := finder.NodeConfig("master").GetInitFile(); got != initScriptContents {
		t.Errorf("Master init script: Got %q, but wanted %q.", got, initScriptContents)
	}
	if got := finder.ClusterComponentObject("kube-apiserver", "pod").GetInlined().GetFields()["biff"].GetStringValue(); got != "bam" {
		t.Errorf("Master kubelet config: Got %q, but wanted %q.", got, "bam")
	}
	if got := finder.ClusterComponentObject("kubelet-config", "kubelet-config-pod").GetInlined().GetFields()["foo"].GetStringValue(); got != "bar" {
		t.Errorf("Master kubelet config: Got %q, but wanted %q.", got, "bar")
	}
}
