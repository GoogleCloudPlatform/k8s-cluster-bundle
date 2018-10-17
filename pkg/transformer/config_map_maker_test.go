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

package transformer

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
  name: zork
`

	out, err := converter.Struct.ProtoToYAML(cfg.cfgMap)
	if err != nil {
		t.Fatalf("error converting struct to yaml: %v", err)
	}

	if string(out) != exp {
		t.Errorf("config map generator; got yaml:\n%s\nbut wanted:\n%s", string(out), exp)
	}

	// make sure it can actually be converted to a K8S ConfigMap
	o, err := converter.FromStruct(cfg.cfgMap).ToConfigMap()
	if err != nil {
		t.Fatalf("error converting struct to config map: %v", err)
	}

	if n := o.ObjectMeta.Name; n != "zork" {
		t.Errorf("got name %s but expected name zork", n)
	}
}
