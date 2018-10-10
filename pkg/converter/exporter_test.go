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
	"strings"
	"testing"

	structpb "github.com/golang/protobuf/ptypes/struct"
)

var (
	objForExport = []string{`
apiVersion: v1
kind: Pod
metadata:
  name: rescheduler
`, `
apiVersion: v1
kind: Pod
metadata:
  name: etcd
`}
)

func TestExporterMulti(t *testing.T) {
	var obj []*structpb.Struct
	for _, o := range objForExport {
		spb, err := Struct.YAMLToProto([]byte(o))
		if err != nil {
			t.Fatalf("Failed to parse yaml:%v \n%s", err, o)
		}
		obj = append(obj, ToStruct(spb))
	}
	exp := ObjectExporter{obj}

	multi, err := exp.ExportAsMultiYAML()
	if err != nil {
		t.Fatalf("Failed to multi-export yaml: %v", err)
	}
	if len(multi) != 2 {
		t.Fatalf("Got items %v, but expected exactly 2", multi)
	}
}

func TestExporterSingle(t *testing.T) {
	var obj []*structpb.Struct
	for _, o := range objForExport {
		spb, err := Struct.YAMLToProto([]byte(o))
		if err != nil {
			t.Fatalf("Failed to parse yaml:%v \n%s", err, o)
		}
		obj = append(obj, ToStruct(spb))
	}
	exp := ObjectExporter{obj}

	single, err := exp.ExportAsYAML()
	if err != nil {
		t.Fatalf("failed to single-export yaml: %v", err)
	}
	if !strings.Contains(single, "\n---\n") {
		t.Errorf("Got %s, but expected yaml to contain document join string", single)
	}
	if !strings.Contains(single, "rescheduler") {
		t.Errorf("Got %s, but expected yaml to contain 'rescheduler'", single)
	}
	if !strings.Contains(single, "etcd") {
		t.Errorf("Got %s, but expected yaml to contain 'etcd'", single)
	}
}
