// Copyright 2019 Google LLC
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

package jsonnet

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/build"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/google/go-jsonnet"
)

func makeComponent(t *testing.T, data string) *bundle.Component {
	t.Helper()

	compBuilder, err := converter.FromYAMLString(data).ToComponentBuilder()
	if err != nil {
		t.Fatal(err)
	}

	path, err := filepath.Abs("builder.yaml")
	if err != nil {
		t.Fatal(err)
	}

	comp, err := build.NewLocalInliner("").ComponentFiles(context.Background(), compBuilder, path)
	if err != nil {
		t.Fatal(err)
	}

	return comp
}

func TestRawStringApplier_MultiItems(t *testing.T) {
	const data = `
apiVersion: bundle.gke.io/v1alpha1
kind: ComponentBuilder
componentName: kubecore
version: 11.0.0
objectFiles:
- url: ../../../examples/jsonnet/pod.builder.yaml
`

	comp := makeComponent(t, data)

	usedParams := map[string]interface{}{
		"DNSPolicy":      "FooBarPolicy",
		"ContainerImage": "MyContainerImage",
	}
	notUsedParams := map[string]interface{}{
		"AnotherDNSPolicy":      "BlooBlarPolicy",
		"AnotherContainerImage": "AnotherContainerImageVal",
	}

	opts := func() options.JSONOptions {
		allMap := map[string]interface{}{}
		for k, v := range usedParams {
			allMap[k] = v
		}
		for k, v := range notUsedParams {
			allMap[k] = v
		}
		return allMap
	}()

	newComp, err := NewApplier(nil).ApplyOptions(comp, opts)
	if err != nil {
		t.Fatalf("Error applying options: %v", err)
	}
	if newComp == nil {
		t.Fatalf("new-component must not be nil")
	}
	if len(newComp.Spec.Objects) == 0 {
		t.Fatalf("no objects found in new component")
	}

	strval, err := (&converter.ObjectExporter{newComp.Spec.Objects}).ExportAsYAML()
	if err != nil {
		t.Fatalf("Error converting objects to yaml: %v", err)
	}

	for _, val := range usedParams {
		vstr := val.(string)
		if !strings.Contains(strval, vstr) {
			t.Errorf("expected object yaml:\n%s\nto contain %q", strval, vstr)
		}
	}
	for _, val := range notUsedParams {
		vstr := val.(string)
		if strings.Contains(strval, vstr) {
			t.Errorf("expected object yaml:\n%s\nto NOT contain %q", strval, vstr)
		}
	}
}

type localImporter struct{}

func (localImporter) Import(importedFrom, importPath string) (jsonnet.Contents, string, error) {
	b, err := ioutil.ReadFile(filepath.Clean(filepath.Join(filepath.Dir(importedFrom), importPath)))
	if err != nil {
		return jsonnet.Contents{}, "", err
	}
	return jsonnet.MakeContents(string(b)), importPath, nil
}

func TestImports(t *testing.T) {
	const data = `
apiVersion: bundle.gke.io/v1alpha1
kind: ComponentBuilder
componentName: kubecore
version: 11.0.0
objectFiles:
- url: ../../../examples/jsonnet/pod-with-imports.builder.yaml
`

	opts := map[string]interface{}{
		"DNSPolicy":      "FooBarPolicy",
		"ContainerImage": "MyContainerImage",
	}

	t.Run("disabled", func(t *testing.T) {
		comp := makeComponent(t, data)

		if _, err := NewApplier(nil).ApplyOptions(comp, opts); err == nil {
			t.Fatalf("Expected error")
		}
	})

	t.Run("custom local", func(t *testing.T) {
		comp := makeComponent(t, data)

		if _, err := NewApplier(localImporter{}).ApplyOptions(comp, opts); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

func TestExample(t *testing.T) {
	example, err := filepath.Abs("../../../examples/jsonnet/builder.yaml")
	if err != nil {
		t.Fatalf("failed to make path: %v", err)
	}

	data, err := ioutil.ReadFile(example)
	if err != nil {
		t.Fatalf("failed to read builder: %v", err)
	}
	compBuilder, err := converter.FromYAML(data).ToComponentBuilder()
	if err != nil {
		t.Fatal(err)
	}

	comp, err := build.NewLocalInliner("").ComponentFiles(context.Background(), compBuilder, example)
	if err != nil {
		t.Fatal(err)
	}

	opts := map[string]interface{}{
		"DNSPolicy":      "FooBarPolicy",
		"ContainerImage": "MyContainerImage",
	}

	if _, err := NewApplier(localImporter{}).ApplyOptions(comp, opts); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
