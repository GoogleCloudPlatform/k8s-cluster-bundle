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
)

const componentDataExample = `
kind: Bundle
components:
- spec:
    componentName: kube-apiserver
    version: 1.0.0
    objects:
    - kind: Pod
      apiVersion: v1
`

func TestDataParse(t *testing.T) {
	data, err := FromYAMLString(componentDataExample).ToBundle()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(data.Components); l == 0 {
		t.Fatalf("Zero components found in the Bundle; expected exactly 1")
	}

	if n := data.Components[0].Spec.ComponentName; n != "kube-apiserver" {
		t.Errorf("Got name %q, expected name %q", n, "kube-apiserver")
	}
}
