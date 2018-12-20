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
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

func TestMakeConfigMap(t *testing.T) {
	cfg := newConfigMapMaker("zork")
	cfg.addData("foo", "bar")
	cfg.addData("biff", "bam")

	exp := `apiVersion: v1
data:
  biff: bam
  foo: bar
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: zork
`

	out, err := cfg.toUnstructured()
	if err != nil {
		t.Fatalf("error converting config map to unstructured: %v", err)
	}

	if n := cfg.cfgMap.ObjectMeta.Name; n != "zork" {
		t.Errorf("got name %s but expected name zork", n)
	}

	if val := cfg.cfgMap.Data["foo"]; val != "bar" {
		t.Errorf("got val %s but expected bar", val)
	}

	s, err := converter.FromObject(out).ToYAML()
	if err != nil {
		t.Errorf("error converting config map to yaml: %v", err)
	}
	if string(s) != exp {
		t.Errorf("Expected serialized yaml\n%s\n but got\n%s", string(s), exp)
	}
}
