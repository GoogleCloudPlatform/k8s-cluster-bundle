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
	"context"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

func TestRealisticDataParseAndInline(t *testing.T) {
	ctx := context.Background()
	b, err := testutil.ReadData("../../", "examples/component-data-files.yaml")
	if err != nil {
		t.Fatalf("Error reading file %v", err)
	}

	dataFiles, err := converter.FromYAML(b).ToComponentData()
	if err != nil {
		t.Fatalf("Error calling ToComponentDataFiles(): %v", err)
	}

	if l := len(dataFiles.ComponentFiles); l == 0 {
		t.Fatalf("found zero files, but expected some")
	}

	pathPrefix := testutil.TestPathPrefix("../../", "examples/component-data-files.yaml")
	inliner := NewLocalInliner(pathPrefix)

	newData, err := inliner.InlineComponentDataFiles(ctx, dataFiles)
	if err != nil {
		t.Fatalf("Error calling InlineComponentDataFiles(): %v", err)
	}

	moreInlined, err := inliner.InlineComponentsInData(ctx, newData)
	if err != nil {
		t.Fatalf("Error calling InlineComponentsInData(): %v", err)
	}

	_, err = converter.FromObject(moreInlined).ToYAML()
	if err != nil {
		t.Fatalf("Error converting the inlined data back into YAML: %v", err)
	}
}
