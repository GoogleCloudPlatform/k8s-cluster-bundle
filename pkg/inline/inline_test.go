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
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
)

const componentData = `
components:
- spec:
    componentName: kube-apiserver
    objectFiles:
    - url: 'path/to/kube_apiserver.yaml'
- spec:
    componentName: kubelet-config
    objectFiles:
    - url: 'path/to/kubelet/config.yaml'
`

type fakeLocalReader struct{}

func (*fakeLocalReader) ReadFileObj(ctx context.Context, file bundle.File) ([]byte, error) {
	url := file.URL
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	switch url {
	case "path/to/kubelet/config.yaml":
		return []byte(`
metadata:
  name: foobar
kind: Zork
apiVersion: v1
foo: bar`), nil

	case "path/to/multidoc.yaml":
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

	case "parent/path/to/kube_apiserver.yaml":
		fallthrough
	case "path/to/kube_apiserver.yaml":
		return []byte(`
metadata:
  name: biffbam
kind: Zork
apiVersion: v1
biff: bam`), nil

	// Same as above but for absolute paths.
	case "/path/to/kube_apiserver.yaml":
		return []byte(`
metadata:
  name: biffbam
kind: Zork
apiVersion: v1
biff: bam-abs`), nil

	case "/path/to/raw-text.yaml":
		fallthrough
	case "parent/path/to/raw-text.yaml":
		fallthrough
	case "path/to/raw-text.yaml":
		return []byte("foobar"), nil

	case "path/to/rawer-text.yaml":
		return []byte("biffbam"), nil

	case "parent/kube_apiserver_component.yaml":
		return []byte(`
spec:
  componentName: kube-apiserver
  objectFiles:
  - url: 'path/to/kube_apiserver.yaml'
  rawTextFiles:
  - name: some-raw-text
    files:
    - url: 'path/to/raw-text.yaml'`), nil

	case "/parent/kube_apiserver_component.yaml":
		return []byte(`
spec:
  componentName: kube-apiserver-abs
  objectFiles:
  - url: '/path/to/kube_apiserver.yaml'
  rawTextFiles:
  - name: some-raw-text
    files:
    - url: 'file:///path/to/raw-text.yaml'`), nil

	default:
		return nil, fmt.Errorf("unexpected file path %q", file.URL)
	}
}

func TestInlineBundle(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentData).ToBundle()
	if err != nil {
		t.Fatalf("Error parsing bundle: %v", err)
	}
	inliner := NewInlinerWithScheme(
		files.FileScheme,
		&fakeLocalReader{},
		DefaultPathRewriter,
	)

	newdata, err := inliner.InlineComponentsInBundle(ctx, data)
	if err != nil {
		t.Fatalf("Error inlining bundle: %v", err)
	}
	for _, c := range newdata.Components {
		if fi := c.Spec.ObjectFiles; len(fi) > 0 {
			t.Errorf("Found files %v, but expected all the cluster object files to be removed.", fi)
		}
	}
	finder := find.NewComponentFinder(newdata.Components)
	got, err := finder.ObjectsFromUniqueComponent("kube-apiserver", core.ObjectRef{Name: "biffbam"})
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}
	if got == nil || len(got) == 0 {
		t.Fatalf("couldn't retrieve component kube-apiserver with object biffbam")
	}

	f := got[0]
	if tf, ok := f.Object["biff"].(string); !ok || tf != "bam" {
		t.Errorf("kube-apiserver: Got %q, but wanted %q.", got, "bam")
	}
	if annot := f.GetAnnotations(); annot == nil || annot[string(bundle.InlineTypeIdentifier)] != string(bundle.KubeObjectInline) {
		t.Errorf("kube-apiserver inline annotation: Got %v, but wanted %q=%q.", annot, bundle.InlineTypeIdentifier, bundle.KubeObjectInline)
	}

	got, err = finder.ObjectsFromUniqueComponent("kubelet-config", core.ObjectRef{Name: "foobar"})
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}
	if got == nil || len(got) == 0 {
		t.Fatalf("couldn't retrieve component kubelet-config with object foobar")
	}
	f = got[0]
	if tf, ok := f.Object["foo"].(string); !ok || tf != "bar" {
		t.Errorf("kube-config: Got %q, but wanted %q.", got, "bar")
	}
	if annot := f.GetAnnotations(); annot == nil || annot[string(bundle.InlineTypeIdentifier)] != string(bundle.KubeObjectInline) {
		t.Errorf("kube-config inline annotation: Got %v, but wanted %q=%q.", annot, bundle.InlineTypeIdentifier, bundle.KubeObjectInline)
	}
}

const componentDataFiles = `
componentFiles:
- url: 'parent/kube_apiserver_component.yaml'
`

func TestInlineDataFiles(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentDataFiles).ToBundle()
	if err != nil {
		t.Fatalf("Error parsing component data: %v", err)
	}
	inliner := NewInlinerWithScheme(
		files.FileScheme,
		&fakeLocalReader{},
		DefaultPathRewriter,
	)

	newdata, err := inliner.InlineBundleFiles(ctx, data)
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
	found, err := finder.UniqueComponentFromName(comp)
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}
	if found == nil {
		t.Fatalf("could not find component %q in data %v", comp, newdata)
	}

	// now try to inline again
	moreInlined, err := inliner.InlineComponentsInBundle(ctx, newdata)
	finder = find.NewComponentFinder(moreInlined.Components)

	ref := core.ObjectRef{Name: "biffbam"}
	foundObj, err := finder.ObjectsFromUniqueComponent(comp, ref)
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}

	if foundObj == nil || len(foundObj) != 1 {
		t.Fatalf("could not find exactly one object in component %q named %v", comp, ref)
	}
	f := foundObj[0]
	foundval, ok := f.Object["biff"].(string)
	if !ok || foundval != "bam" {
		t.Errorf("found %v, %t but expected value %q", foundval, ok, "bam")
	}
	if annot := f.GetAnnotations(); annot == nil || annot[string(bundle.InlineTypeIdentifier)] != string(bundle.KubeObjectInline) {
		t.Errorf("biffbam inline annotation: Got %v, but wanted %q=%q.", annot, bundle.InlineTypeIdentifier, bundle.KubeObjectInline)
	}

	ref = core.ObjectRef{Name: "some-raw-text"}
	foundObj, err = finder.ObjectsFromUniqueComponent(comp, ref)
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}
	if foundObj == nil || len(foundObj) != 1 {
		t.Fatalf("could not find exactly one object in component %q named %v", comp, ref)
	}
	f = foundObj[0]
	if annot := f.GetAnnotations(); annot == nil || annot[string(bundle.InlineTypeIdentifier)] != string(bundle.RawStringInline) {
		t.Errorf("some-raw-text inline annotation: Got %v, but wanted %q=%q.", annot, bundle.InlineTypeIdentifier, bundle.RawStringInline)
	}
}

const componentsWithMultidoc = `
components:
- spec:
    componentName: multidoc
    objectFiles:
    - url: 'path/to/multidoc.yaml'
`

func TestMultiDoc(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentsWithMultidoc).ToBundle()
	if err != nil {
		t.Fatalf("Error parsing component data: %v", err)
	}
	inliner := NewInlinerWithScheme(
		files.FileScheme,
		&fakeLocalReader{},
		DefaultPathRewriter,
	)

	newdata, err := inliner.InlineComponentsInBundle(ctx, data)
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
	found, err := finder.UniqueComponentFromName(comp)
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}

	if found == nil {
		t.Fatalf("could not find componont %q", comp)
	}

	objname := "biffbam"
	got, err := finder.ObjectsFromUniqueComponent(comp, core.ObjectRef{Name: objname})
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}
	if l := len(got); l != 1 {
		t.Fatalf("got %d objects with name %q, wantedexactly one", l, objname)
	}
	if fv, ok := got[0].Object["biff"].(string); !ok || fv != "bam" {
		t.Errorf("multidoc object: Got %q for key biff, but wanted %q.", fv, "bam")
	}
	objname = "foobar"
	got, err = finder.ObjectsFromUniqueComponent(comp, core.ObjectRef{Name: objname})
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}

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
    componentName: textdoc
    rawTextFiles:
    - name: raw-collection
      files:
      - url: 'path/to/raw-text.yaml'
      - url: 'path/to/rawer-text.yaml'
`

func TestBundleRawText(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentWithText).ToBundle()
	if err != nil {
		t.Fatalf("Error parsing component data: %v", err)
	}
	inliner := NewInlinerWithScheme(
		files.FileScheme,
		&fakeLocalReader{},
		DefaultPathRewriter,
	)

	newdata, err := inliner.InlineComponentsInBundle(ctx, data)
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

	// Validate the raw-collection data blob
	name := "raw-collection"
	fname := "raw-text.yaml"
	comp := "textdoc"
	objs, err := finder.ObjectsFromUniqueComponent(comp, core.ObjectRef{Name: name})
	if err != nil {
		t.Fatalf("unexpected duplicate object error: %v", err)
	}

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
	val, ok := dataMap[fname].(string)
	if !ok {
		t.Fatalf("Could not find text object with key %q value %q for comp %q in object %q", fname, "foobar", comp, name)
	}
	if val != "foobar" {
		t.Fatalf("Got value %q for key %q but expected value %q for comp %q in object %q", val, fname, "foobar", comp, name)
	}

	// Validate the rawer-text data blob
	fname = "rawer-text.yaml"
	comp = "textdoc"
	objs, err = finder.ObjectsFromUniqueComponent(comp, core.ObjectRef{Name: name})
	if err != nil {
		t.Fatalf("unexpected duplicate object error: %v", err)
	}
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
	val, ok = dataMap[fname].(string)
	if !ok || val != "biffbam" {
		t.Fatalf("Could not find text object with key %q value %q for comp %q in object %q", fname, "biffbam", comp, name)
	}
}

const componentDataFilesAbsPath = `
componentFiles:
- url: 'file:///parent/kube_apiserver_component.yaml'
`

func TestInlineDataFiles_AbsolutePath(t *testing.T) {
	ctx := context.Background()
	data, err := converter.FromYAMLString(componentDataFilesAbsPath).ToBundle()
	if err != nil {
		t.Fatalf("Error parsing component data with absolute path: %v", err)
	}
	inliner := NewInlinerWithScheme(
		files.FileScheme,
		&fakeLocalReader{},
		DefaultPathRewriter,
	)

	newdata, err := inliner.InlineBundleFiles(ctx, data)
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
	comp := "kube-apiserver-abs"
	found, err := finder.UniqueComponentFromName(comp)
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}
	if found == nil {
		t.Fatalf("could not find component %q in data %v", comp, newdata)
	}

	// now try to inline again
	moreInlined, err := inliner.InlineComponentsInBundle(ctx, newdata)
	finder = find.NewComponentFinder(moreInlined.Components)

	ref := core.ObjectRef{Name: "biffbam"}
	foundObj, err := finder.ObjectsFromUniqueComponent(comp, ref)
	if err != nil {
		t.Fatalf("unexpected duplicate component error: %v", err)
	}

	if foundObj == nil || len(foundObj) != 1 {
		t.Fatalf("could not find exactly one object in component %q named %v", comp, ref)
	}
	f := foundObj[0]
	foundval, ok := f.Object["biff"].(string)
	if !ok || foundval != "bam-abs" {
		t.Errorf("found %v, %t but expected value %q", foundval, ok, "bam-abs")
	}
}
