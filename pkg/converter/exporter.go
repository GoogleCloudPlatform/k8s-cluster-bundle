package converter

import (
	"strings"

	structpb "github.com/golang/protobuf/ptypes/struct"
)

// ObjectExporter exports cluster objects
type ObjectExporter struct {
	objects []*structpb.Struct
}

// ExportAsMultiYAML converts cluster objects into multiple YAML files.
func (e *ObjectExporter) ExportAsMultiYAML() ([]string, error) {
	var out []string
	var empty []string
	for _, o := range e.objects {
		yaml, err := Struct.ProtoToYAML(o)
		if err != nil {
			return empty, err
		}
		out = append(out, string(yaml))
	}
	return out, nil
}

// ExportAsYAML converts cluster objects into single YAML file.
func (e *ObjectExporter) ExportAsYAML() (string, error) {
	numElements := len(e.objects)
	var sb strings.Builder
	for i, o := range e.objects {
		yaml, err := Struct.ProtoToYAML(o)
		if err != nil {
			return "", err
		}
		sb.Write(yaml)
		if i < numElements-1 {
			// Join the objects into one document.
			sb.WriteString("\n---\n")
		}
	}
	return sb.String(), nil
}
