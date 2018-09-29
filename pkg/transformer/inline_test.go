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
  - metadata:
      name: master
    externalInitFile:
      url: 'file://path/to/init-script.sh'
  components:
  - metadata:
      name: kube-apiserver
    clusterObjectFiles:
    - url: 'file://path/to/kube_apiserver.yaml'
  - metadata:
      name: kubelet-config
    clusterObjectFiles:
    - url: 'file://path/to/kubelet/config.yaml'
`

const (
	initScriptFile    = "path/to/init-script.sh"
	kubeletConfigFile = "path/to/kubelet/config.yaml"
	kubeAPIServerFile = "path/to/kube_apiserver.yaml"

	parentInitScriptFile    = "parent/path/to/init-script.sh"
	parentKubeAPIServerFile = "parent/path/to/kube_apiserver.yaml"

	nodeCfgFile = "parent/some_node_config.yaml"
	compFile    = "parent/kube_apiserver_component.yaml"

	initScriptContents = "#!/bin/bash\necho foo"
)

type fakeLocalReader struct{}

func (*fakeLocalReader) ReadFilePB(ctx context.Context, file *bpb.File) ([]byte, error) {
	url := file.GetUrl()
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	switch url {
	case parentInitScriptFile:
		fallthrough
	case initScriptFile:
		return []byte(initScriptContents), nil

	case kubeletConfigFile:
		return []byte("{\"metadata\": { \"name\": \"foobar\"}, \"foo\": \"bar\"}"), nil

	case parentKubeAPIServerFile:
		fallthrough
	case kubeAPIServerFile:
		return []byte("{\"metadata\": { \"name\": \"biffbam\"}, \"biff\": \"bam\"}"), nil

	case nodeCfgFile:
		return []byte(`
metadata:
  name: master
externalInitFile:
  url: 'file://path/to/init-script.sh'`), nil

	case compFile:
		return []byte(`
metadata:
  name: kube-apiserver
clusterObjectFiles:
- url: 'file://path/to/kube_apiserver.yaml'`), nil

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
		Reader: &fakeLocalReader{},
	}

	newpb, err := inliner.Inline(ctx, bp, &InlineOptions{})
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
	if got := finder.ClusterComponentObject("kube-apiserver", "biffbam")[0].GetFields()["biff"].GetStringValue(); got != "bam" {
		t.Errorf("Master kubelet config: Got %q, but wanted %q.", got, "bam")
	}
	if got := finder.ClusterComponentObject("kubelet-config", "foobar")[0].GetFields()["foo"].GetStringValue(); got != "bar" {
		t.Errorf("Master kubelet config: Got %q, but wanted %q.", got, "bar")
	}
}

const twoLayerBundle = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: test-bundle
spec:
  nodeConfigFiles:
  - url: 'file://parent/some_node_config.yaml'
  componentFiles:
  - url: 'file://parent/kube_apiserver_component.yaml'
`

func TestTwoLayerInline(t *testing.T) {
	ctx := context.Background()
	b, err := converter.Bundle.YAMLToProto([]byte(twoLayerBundle))
	if err != nil {
		t.Fatalf("Error parsing bundle: %v", err)
	}
	bp := converter.ToBundle(b)
	inliner := &Inliner{
		Reader: &fakeLocalReader{},
	}

	newpb, err := inliner.Inline(ctx, bp, &InlineOptions{})
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
	found := finder.ClusterComponentObject("kube-apiserver", "biffbam")
	comp := finder.ClusterComponent("kube-apiserver")
	if len(found) == 0 {
		t.Fatalf("could not find component %q and object %q in object %v", "kube-apiserver", "biffbam", comp)
	}
	if got := found[0].GetFields()["biff"].GetStringValue(); got != "bam" {
		t.Errorf("Master kubelet config: Got %q, but wanted %q.", got, "bam")
	}

	// Now try with only inlining the top-layer
	toppb, err := inliner.Inline(ctx, bp, &InlineOptions{TopLayerOnly: true})
	if err != nil {
		t.Fatalf("Error inlining bundle: %v", err)
	}
	finder, err = find.NewBundleFinder(toppb)
	if err != nil {
		t.Fatalf("Error creating bundle finder: %v", err)
	}
	comp = finder.ClusterComponent("kube-apiserver")
	if len(found) == 0 {
		t.Fatalf("could not find component %q and object %q in object %v", "kube-apiserver", "biffbam", comp)
	}
	found = finder.ClusterComponentObject("kube-apiserver", "biffbam")
	if len(found) != 0 {
		t.Fatalf("found %v but did not expect to find anything", found)
	}
}
