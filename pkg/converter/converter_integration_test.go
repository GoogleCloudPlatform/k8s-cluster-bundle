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
	"os"
	"testing"
)

func TestRealisticDataParse_BundleBuilder(t *testing.T) {
	b, err := os.ReadFile("../../examples/cluster/bundle-builder-example.yaml")
	if err != nil {
		t.Fatalf("Error reading file %v", err)
	}

	dataFiles, err := FromYAML(b).ToBundleBuilder()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(dataFiles.ComponentFiles); l == 0 {
		t.Fatalf("found zero files, but expected some")
	}
}

func TestRealisticDataParse_ComponentSet(t *testing.T) {
	b, err := os.ReadFile("../../examples/cluster/component-set.yaml")
	if err != nil {
		t.Fatalf("Error reading file %v", err)
	}

	cset, err := FromYAML(b).ToComponentSet()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(cset.Spec.Components); l == 0 {
		t.Fatalf("found zero components, but expected some")
	}
}

func TestRealisticDataParse_ComponentBuilder(t *testing.T) {
	b, err := os.ReadFile("../../examples/component/etcd-component-builder.yaml")
	if err != nil {
		t.Fatalf("Error reading file %v", err)
	}

	comp, err := FromYAML(b).ToComponentBuilder()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(comp.ObjectFiles); l == 0 {
		t.Fatalf("found zero object urls, but expected some")
	}
}

func TestRealisticDataParse_Component(t *testing.T) {
	b, err := os.ReadFile("../../examples/component/etcd-component.yaml")
	if err != nil {
		t.Fatalf("Error reading file %v", err)
	}

	comp, err := FromYAML(b).ToComponent()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(comp.Spec.Objects); l == 0 {
		t.Fatalf("found zero objects, but expected some")
	}
}
