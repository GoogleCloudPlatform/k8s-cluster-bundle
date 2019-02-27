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

package build

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validate"
)

func TestRealisticDataParseAndInline_Bundle(t *testing.T) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("../../examples/cluster/bundle-builder-example.yaml")
	if err != nil {
		t.Fatalf("Error reading file %v", err)
	}

	dataFiles, err := converter.FromYAML(b).ToBundleBuilder()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(dataFiles.ComponentFiles); l == 0 {
		t.Fatalf("found zero files, but expected some")
	}

	inliner := NewLocalInliner("../../examples/cluster/")

	moreInlined, err := inliner.BundleFiles(ctx, dataFiles)
	if err != nil {
		t.Fatalf("Error calling BundleFiles(): %v", err)
	}

	_, err = converter.FromObject(moreInlined).ToYAML()
	if err != nil {
		t.Fatalf("Error converting the inlined data back into YAML: %v", err)
	}

	// Ensure it validates
	if errs := validate.Components(moreInlined.Components); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Errors in validaton: %q", e.Error())
		}
	}
}

func TestRealisticDataParseAndInline_Component(t *testing.T) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("../../examples/component/etcd-component-builder.yaml")
	if err != nil {
		t.Fatalf("Error reading file %v", err)
	}

	cb, err := converter.FromYAML(b).ToComponentBuilder()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(cb.ObjectFiles); l == 0 {
		t.Fatalf("found zero files, but expected some")
	}

	inliner := NewLocalInliner("../../examples/component/")

	component, err := inliner.ComponentFiles(ctx, cb)
	if err != nil {
		t.Fatalf("Error calling ComponentFiles(): %v", err)
	}

	yaml, err := converter.FromObject(component).ToYAML()
	if err != nil {
		t.Fatalf("Error converting the inlined component back into YAML: %v", err)
	}

	// Ensure it validates
	if errs := validate.Component(component); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Errors in validaton: %q", e.Error())
		}
	}
}
