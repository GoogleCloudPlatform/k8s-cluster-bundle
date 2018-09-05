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

const crdYaml = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.example.com
spec:
  group: stable.example.com
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: crontabs
    singular: crontab
    kind: CronTab
    shortNames:
    - ct
`

func TestToCRD(t *testing.T) {
	s, err := Struct.YAMLToProto([]byte(crdYaml))
	if err != nil {
		t.Fatalf("Error converting yaml to struct: %v", err)
	}
	spb := ToStruct(s)

	crd, err := FromStruct(spb).ToCRD()
	if err != nil {
		t.Fatalf("Error converting struct to crd: %v", err)
	}

	if crd.Kind != "CustomResourceDefinition" {
		t.Fatalf("Got kind %q, but expected CustomResourceDefinition", crd.Kind)
	}
}
