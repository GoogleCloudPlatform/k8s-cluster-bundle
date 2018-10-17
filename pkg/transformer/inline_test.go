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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
)

const bundleWithRefs = `
apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: test-bundle
spec:
  components:
  - metadata:
      name: kube-apiserver
    spec:
      clusterObjectFiles:
      - url: 'file://path/to/kube_apiserver.yaml'
  - metadata:
      name: kubelet-config
    spec:
      clusterObjectFiles:
      - url: 'file://path/to/kubelet/config.yaml'
`

const (
	kubeletConfigFile = "path/to/kubelet/config.yaml"
	kubeAPIServerFile = "path/to/kube_apiserver.yaml"
	multiDocFile      = "path/to/multidoc.yaml"

	parentInitScriptFile    = "parent/path/to/init-script.sh"
	parentKubeAPIServerFile = "parent/path/to/kube_apiserver.yaml"

	compFile = "parent/kube_apiserver_component.yaml"

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

	case kubeletConfigFile:
		return []byte("{\"metadata\": { \"name\": \"foobar\"}, \"foo\": \"bar\"}"), nil

	case multiDocFile:
		return []byte(`
metadata:
  name: foobar
foo: bar
---
metadata:
  name: biffbam
biff: bam`), nil

	case parentKubeAPIServerFile:
		fallthrough
	case kubeAPIServerFile:
		return []byte("{\"metadata\": { \"name\": \"biffbam\"}, \"biff\": \"bam\"}"), nil

	case "path/to/raw-text.yaml":
		return []byte("foobar"), nil

	case "path/to/rawer-text.yaml":
		return []byte("biffbam"), nil

	case compFile:
		return []byte(`
metadata:
  name: kube-apiserver
spec:
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
	for _, c := range newpb.GetSpec().GetComponents() {
		if fi := c.GetSpec().ClusterObjectFiles; len(fi) > 0 {
			t.Errorf("Found files %v, but expected all the cluster object files to be removed.", fi)
		}
	}

	if got := finder.ClusterObjects("kube-apiserver", core.ObjectRef{Name: "biffbam"})[0].GetFields()["biff"].GetStringValue(); got != "bam" {
		t.Errorf("Master kubelet config: Got %q, but wanted %q.", got, "bam")
	}
	if got := finder.ClusterObjects("kubelet-config", core.ObjectRef{Name: "foobar"})[0].GetFields()["foo"].GetStringValue(); got != "bar" {
		t.Errorf("Master kubelet config: Got %q, but wanted %q.", got, "bar")
	}
}

const twoLayerBundle = `
apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: test-bundle
spec:
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
	found := finder.ClusterObjects("kube-apiserver", core.ObjectRef{Name: "biffbam"})
	comp := finder.ComponentPackage("kube-apiserver")
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
	comp = finder.ComponentPackage("kube-apiserver")
	if len(found) == 0 {
		t.Fatalf("could not find component %q and object %q in object %v", "kube-apiserver", "biffbam", comp)
	}
	found = finder.ClusterObjects("kube-apiserver", core.ObjectRef{Name: "biffbam"})
	if len(found) != 0 {
		t.Fatalf("found %v but did not expect to find anything", found)
	}
}

const bundleWithMultidoc = `
apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: test-bundle
spec:
  components:
  - metadata:
      name: multidoc
    spec:
      clusterObjectFiles:
      - url: 'file://path/to/multidoc.yaml'
`

func TestMultiDoc(t *testing.T) {
	ctx := context.Background()

	b, err := converter.Bundle.YAMLToProto([]byte(bundleWithMultidoc))
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

	for _, c := range newpb.GetSpec().GetComponents() {
		if fi := c.GetSpec().ClusterObjectFiles; len(fi) > 0 {
			t.Errorf("Found files %v, but expected all the cluster object files to be removed.", fi)
		}
	}

	if got := finder.ClusterObjects("multidoc", core.ObjectRef{Name: "biffbam"})[0].GetFields()["biff"].GetStringValue(); got != "bam" {
		t.Errorf("multidoc object: Got %q, but wanted %q.", got, "bam")
	}
	if got := finder.ClusterObjects("multidoc", core.ObjectRef{Name: "foobar"})[0].GetFields()["foo"].GetStringValue(); got != "bar" {
		t.Errorf("multidoc object: Got %q, but wanted %q.", got, "bar")
	}
}

const bundleWithText = `
apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: test-bundle
spec:
  components:
  - metadata:
      name: textdoc
    spec:
      rawTextFiles:
      - url: 'file://path/to/raw-text.yaml'
      - url: 'file://path/to/rawer-text.yaml'
`

func TestBundleRawText(t *testing.T) {
	ctx := context.Background()

	b, err := converter.Bundle.YAMLToProto([]byte(bundleWithText))
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

	outBun, err := converter.Bundle.ProtoToYAML(newpb)
	if err != nil {
		t.Fatalf("error converting inlined bundle back into a bundle-yaml: %v", err)
	}

	for _, c := range newpb.GetSpec().GetComponents() {
		if fi := c.GetSpec().RawTextFiles; len(fi) > 0 {
			t.Errorf("Found files %v, but expected all the raw text files to be removed.", fi)
		}
	}

	name := "raw-text.yaml"
	objs := finder.ClusterObjects("textdoc", core.ObjectRef{Name: name})
	if len(objs) == 0 {
		t.Fatalf("Couldn't find cluster object in named %s in bundle %s", name, string(outBun))
	}
	if got := objs[0].GetFields()["data"].GetStructValue().GetFields()["raw-text.yaml"].GetStringValue(); got != "foobar" {
		t.Errorf("textdoc object: Got %q, but wanted foobar.", got)
	}

	name = "rawer-text.yaml"
	objs = finder.ClusterObjects("textdoc", core.ObjectRef{Name: name})
	if len(objs) == 0 {
		t.Fatalf("Couldn't find cluster object in named %s in bundle %s", name, string(outBun))
	}
	if got := objs[0].GetFields()["data"].
		GetStructValue().GetFields()["rawer-text.yaml"].GetStringValue(); got != "biffbam" {
		t.Errorf("textdoc object: Got %q, but wanted biffbam", got)
	}
}
