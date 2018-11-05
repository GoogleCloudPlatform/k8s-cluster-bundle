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

package inline

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

const componentData = `
components:
- spec:
    canonicalName: kube-apiserver
    objectFiles:
    - url: 'file://path/to/kube_apiserver.yaml'
- spec:
    canonicalName: kubelet-config
    objectFiles:
    - url: 'file://path/to/kubelet/config.yaml'
`

const (
	kubeletConfigFile = "path/to/kubelet/config.yaml"
	kubeAPIServerFile = "path/to/kube_apiserver.yaml"
	multiDocFile      = "path/to/multidoc.yaml"

	parentKubeAPIServerFile = "parent/path/to/kube_apiserver.yaml"
)

type fakeLocalReader struct{}

func (*fakeLocalReader) ReadFileObj(ctx context.Context, file bpb.File) ([]byte, error) {
	url := file.URL
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	switch url {
	case kubeletConfigFile:
		return []byte(`
metadata:
  name: foobar
kind: Zork
apiVersion: v1
foo: bar`), nil

	case multiDocFile:
		return []byte(`
metadata:
  name: foobar
kind: Zork
apiVersion: v1
foo: bar
---
metadata:
  name: biffbam
kind: Zork
apiVersion: v1
biff: bam`), nil

	case parentKubeAPIServerFile:
		fallthrough
	case kubeAPIServerFile:
		return []byte(`
metadata:
  name: biffbam
kind: Zork
apiVersion: v1
biff: bam`), nil

	case "parent/path/to/raw-text.yaml":
		fallthrough
	case "path/to/raw-text.yaml":
		return []byte("foobar"), nil

	case "path/to/rawer-text.yaml":
		return []byte("biffbam"), nil

	case "parent/kube_apiserver_component.yaml":
		return []byte(`
spec:
  canonicalName: kube-apiserver
  objectFiles:
  - url: 'file://path/to/kube_apiserver.yaml'
  rawTextFiles:
  - url: 'file://path/to/raw-text.yaml'`), nil

	default:
		return nil, fmt.Errorf("unexpected file path %q", file.URL)
	}
}

func TestInlineBundle(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentData).ToComponentData()
	if err != nil {
		t.Fatalf("Error parsing bundle: %v", err)
	}
	inliner := &Inliner{
		Reader: &fakeLocalReader{},
	}

	newdata, err := inliner.InlineComponentsInData(ctx, data)
	if err != nil {
		t.Fatalf("Error inlining bundle: %v", err)
	}
	for _, c := range newdata.Components {
		if fi := c.Spec.ObjectFiles; len(fi) > 0 {
			t.Errorf("Found files %v, but expected all the cluster object files to be removed.", fi)
		}
	}

	finder := find.NewComponentFinder(newdata.Components)
	got := finder.ObjectsFromUniqueComponent("kube-apiserver", core.ObjectRef{Name: "biffbam"})
	if got == nil || len(got) == 0 {
		t.Fatalf("couldn't retrieve component kube-apiserver with object biffbam")
	}

	f := got[0].Object["biff"]
	if tf, ok := f.(string); !ok || tf != "bam" {
		t.Errorf("kube-apiserver: Got %q, but wanted %q.", got, "bam")
	}

	got = finder.ObjectsFromUniqueComponent("kubelet-config", core.ObjectRef{Name: "foobar"})
	if got == nil || len(got) == 0 {
		t.Fatalf("couldn't retrieve component kubelet-config with object foobar")
	}
	f = got[0].Object["foo"]
	if tf, ok := f.(string); !ok || tf != "bar" {
		t.Errorf("kube-apiserver: Got %q, but wanted %q.", got, "bar")
	}
}

const componentDataFiles = `
componentFiles:
- url: 'file://parent/kube_apiserver_component.yaml'
`

func TestInlineDataFiles(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentDataFiles).ToComponentData()
	if err != nil {
		t.Fatalf("Error parsing component data: %v", err)
	}
	inliner := &Inliner{
		Reader: &fakeLocalReader{},
	}

	newdata, err := inliner.InlineComponentDataFiles(ctx, data)
	if err != nil {
		t.Fatalf("Error inlining component data files: %v", err)
	}
	if len(newdata.Components) != 1 {
		t.Fatalf("found components %v but expected exactly one", newdata.Components)
	}
	if len(newdata.Components[0].Spec.Objects) != 0 {
		t.Fatalf("found unexpected object files: %v", newdata.Components[0].Spec.Objects)
	}

	finder := find.NewComponentFinder(newdata.Components)
	comp := "kube-apiserver"
	if found := finder.UniqueComponentFromName(comp); found == nil {
		t.Fatalf("could not find component %q in data %v", comp, newdata)
	}

	// now try to inline again
	moreInlined, err := inliner.InlineComponentsInData(ctx, newdata)
	finder = find.NewComponentFinder(moreInlined.Components)

	ref := core.ObjectRef{Name: "biffbam"}
	found := finder.ObjectsFromUniqueComponent(comp, ref)

	if found == nil || len(found) != 1 {
		t.Fatalf("could not find exactly one object in component %q named %v", comp, ref)
	}
	foundval, ok := found[0].Object["biff"].(string)
	if !ok || foundval != "bam" {
		t.Errorf("found %v, %t but expected value %q", foundval, ok, "bam")
	}

	ref = core.ObjectRef{Name: "raw-text.yaml"}
	found = finder.ObjectsFromUniqueComponent(comp, ref)
	if found == nil || len(found) != 1 {
		t.Fatalf("could not find exactly one object in component %q named %v", comp, ref)
	}
}

const componentsWithMultidoc = `
components:
- spec:
    canonicalName: multidoc
    objectFiles:
    - url: 'file://path/to/multidoc.yaml'
`

func TestMultiDoc(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentsWithMultidoc).ToComponentData()
	if err != nil {
		t.Fatalf("Error parsing component data: %v", err)
	}
	inliner := &Inliner{Reader: &fakeLocalReader{}}

	newdata, err := inliner.InlineComponentsInData(ctx, data)
	if err != nil {
		t.Fatalf("Error inlining component data: %v", err)
	}

	finder := find.NewComponentFinder(newdata.Components)
	for _, c := range newdata.Components {
		if fi := c.Spec.ObjectFiles; len(fi) > 0 {
			t.Errorf("Found files %v, but expected all the object files to be removed.", fi)
		}
	}

	comp := "multidoc"
	if found := finder.UniqueComponentFromName(comp); found == nil {
		t.Fatalf("could not find componont %q", comp)
	}

	objname := "biffbam"
	got := finder.ObjectsFromUniqueComponent(comp, core.ObjectRef{Name: objname})
	if l := len(got); l != 1 {
		t.Fatalf("got %d objects with name %q, wantedexactly one", l, objname)
	}
	if fv, ok := got[0].Object["biff"].(string); !ok || fv != "bam" {
		t.Errorf("multidoc object: Got %q for key biff, but wanted %q.", fv, "bam")
	}
	objname = "foobar"
	got = finder.ObjectsFromUniqueComponent(comp, core.ObjectRef{Name: objname})
	if l := len(got); l != 1 {
		t.Fatalf("got %d objects with name %q, wantedexactly one", l, objname)
	}
	if fv, ok := got[0].Object["foo"].(string); !ok || fv != "bar" {
		t.Errorf("multidoc object: Got %q for key foo, but wanted %q.", fv, "bar")
	}
}

const componentWithText = `
components:
- spec:
    canonicalName: textdoc
    rawTextFiles:
    - url: 'file://path/to/raw-text.yaml'
    - url: 'file://path/to/rawer-text.yaml'
`

func TestBundleRawText(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentWithText).ToComponentData()
	if err != nil {
		t.Fatalf("Error parsing component data: %v", err)
	}
	inliner := &Inliner{Reader: &fakeLocalReader{}}

	newdata, err := inliner.InlineComponentsInData(ctx, data)
	if err != nil {
		t.Fatalf("Error inlining components in data: %v", err)
	}
	finder := find.NewComponentFinder(newdata.Components)

	_, err = converter.FromObject(newdata).ToYAML()
	if err != nil {
		t.Fatalf("error converting inlined bundle back into a bundle-yaml: %v", err)
	}

	for _, c := range newdata.Components {
		if fi := c.Spec.RawTextFiles; len(fi) > 0 {
			t.Errorf("Found files %v, but expected all the raw text files to be removed.", fi)
		}
	}

	name := "raw-text.yaml"
	comp := "textdoc"
	objs := finder.ObjectsFromUniqueComponent(comp, core.ObjectRef{Name: name})
	if l := len(objs); l != 1 {
		t.Fatalf("found %d objects, but expected exactly 1 for component %q with object name %s ", l, comp, name)
	}
	o := objs[0].Object
	if len(o) == 0 || o["data"] == nil {
		t.Fatalf("textdoc object for comp %q in object %q: could not access data field", comp, name)
	}
	dataObj := objs[0].Object["data"]
	dataMap, ok := dataObj.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a map of string to interface for comp %q in object %q", comp, name)
	}
	val, ok := dataMap[name].(string)
	if !ok || val != "foobar" {
		t.Fatalf("Could not find text object with key %q value %q for comp %q in object %q", name, "foobar", comp, name)
	}

	name = "rawer-text.yaml"
	comp = "textdoc"
	objs = finder.ObjectsFromUniqueComponent(comp, core.ObjectRef{Name: name})
	if l := len(objs); l != 1 {
		t.Fatalf("found %d objects, but expected exactly 1 for component %q with object name %s ", l, comp, name)
	}
	o = objs[0].Object
	if len(o) == 0 || o["data"] == nil {
		t.Fatalf("textdoc object for comp %q in object %q: could not access data field", comp, name)
	}
	dataObj = objs[0].Object["data"]
	dataMap, ok = dataObj.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a map of string to interface for comp %q in object %q", comp, name)
	}
	val, ok = dataMap[name].(string)
	if !ok || val != "biffbam" {
		t.Fatalf("Could not find text object with key %q value %q for comp %q in object %q", name, "biffbam", comp, name)
	}
}
