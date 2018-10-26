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
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

func TestRealisticBundleParse(t *testing.T) {
	bundleContents, err := testutil.ReadTestBundle("../testutil/testdata")
	if err != nil {
		t.Fatalf("Error reading bundle file %v", err)
	}
	b, err := Bundle.YAMLToProto(bundleContents)
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}
	bp := ToBundle(b)

	expID := "1.9.7.testbundle-zork"
	if bp.GetMetadata().GetName() != expID {
		t.Errorf("Got name %q, expected name %q", bp.GetMetadata().GetName(), expID)
	}
}

func TestRealisticBundleParseK8sBundle(t *testing.T) {
	bundleContents, err := testutil.ReadTestBundle("../testutil/testdata")
	if err != nil {
		t.Fatalf("Error reading bundle file %v", err)
	}
	bp, err := YAMLToK8sBundle(bundleContents)
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	expID := "1.9.7.testbundle-zork"
	if bp.ObjectMeta.Name != expID {
		t.Errorf("Got name %q, expected name %q", bp.ObjectMeta.Name, expID)
	}
}
