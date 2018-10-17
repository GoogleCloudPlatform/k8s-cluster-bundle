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
