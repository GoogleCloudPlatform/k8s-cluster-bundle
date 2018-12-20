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
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validate"
)

func TestRealisticDataParseAndInline(t *testing.T) {
	ctx := context.Background()
	b, err := testutil.ReadData("../../", "examples/cluster/bundle-builder-example.yaml")
	if err != nil {
		t.Fatalf("Error reading file %v", err)
	}

	dataFiles, err := converter.FromYAML(b).ToBundleBuilder()
	if err != nil {
		t.Fatalf("error converting data: %v", err)
	}

	if l := len(dataFiles.ComponentFiles); l == 0 {
		t.Fatalf("found zero files, but expected some")
	}

	pathPrefix := testutil.TestPathPrefix("../../", "examples/cluster/bundle-builder-example.yaml")
	inliner := NewLocalInliner(pathPrefix)

	moreInlined, err := inliner.BundleFiles(ctx, dataFiles)
	if err != nil {
		t.Fatalf("Error calling BundleFiles(): %v", err)
	}

	_, err = converter.FromObject(moreInlined).ToYAML()
	if err != nil {
		t.Fatalf("Error converting the inlined data back into YAML: %v", err)
	}

	// Ensure it validates
	if errs := validate.AllComponents(moreInlined.Components); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Errors in validaton: %q", e.Error())
		}
	}
}
