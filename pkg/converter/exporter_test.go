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
