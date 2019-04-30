package build

import (
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"testing"
)

func BenchmarkBuildAndPatch_Component(t *testing.B) {
	component := `
kind: Component
spec:
  objects:
  - apiVersion: v1
    kind: Pod
  - kind: PatchTemplateBuilder
    apiVersion: bundle.gke.io/v1alpha1
    buildSchema:
      required:
        - Namespace
      properties:
        Namespace:
          type: string
    targetSchema:
      required:
      - PodName
      properties:
        PodName:
          type: string
    template: |
      kind: Pod
      metadata:
        namespace: {{.Namespace}}
        name: {{.PodName}}`

	for i := 0; i < t.N; i++ {
		c, err := converter.FromYAMLString(component).ToComponent()
		if err != nil {
			panic("error parsing component")
		}
		newComp, err := ComponentPatchTemplates(c, &filter.Options{}, map[string]interface{}{
			"Namespace": "foo",
		})
		if err != nil {
			panic("error patching component")
		}
		_, err = converter.FromObject(newComp).ToYAML()
		if err != nil {
			panic("error converting object")
		}
	}

}
