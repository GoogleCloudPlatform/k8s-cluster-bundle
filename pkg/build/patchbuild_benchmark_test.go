

package build

import (
  //"context"
 // "io/ioutil"
  "testing"

  "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
  "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"

 // "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validate"
)
/*
      opts: map[string]interface{}{
        "Namespace": "foo",
      },
      component: 


          c, err := converter.FromYAMLString(tc.component).ToComponent()
      if err != nil {
        t.Fatalf("Error converting component %s: %v", tc.component, err)
      }

      newComp, err := ComponentPatchTemplates(c, tc.customFilter, tc.opts)
      cerr := testutil.CheckErrorCases(err, tc.expErrSubstr)
      if cerr != nil {
        t.Error(cerr)
      }
      if err != nil {
        // We hit an expected error, but we can't continue on because newComp is nil.
        return
      }

      compBytes, err := converter.FromObject(newComp).ToYAML()
      if err != nil {
        t.Fatalf("Error converting back to yaml: %v", err)
      }

      compStr := strings.Trim(string(compBytes), " \n\r")
      expStr := strings.Trim(tc.output, " \n\r")
      if expStr != compStr {
        t.Errorf("got yaml\n%s\n\nbut expected output yaml to be\n%s", compStr, expStr)
      }
    })
*/
func BenchmarkBuildAndPatch_Component(t *testing.B) {
 /* b, _ := ioutil.ReadFile("../../examples/component/etcd-component-builder.yaml")
  dataPath := "../../examples/component/etcd-component-builder.yaml"

  for i := 0; i < t.N; i++ {
    cb, _ := converter.FromYAML(b).ToComponentBuilder()
    inliner := NewLocalInliner("../../examples/component/")
    component, _:= inliner.ComponentFiles(context.Background(), cb, dataPath)
    _, _ = converter.FromObject(component).ToYAML()
    validate.Component(component) 
  }*/

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
      c, _ := converter.FromYAMLString(component).ToComponent()
      newComp, _ := ComponentPatchTemplates(c, &filter.Options{}, map[string]interface{}{
        "Namespace": "foo",
      })
      converter.FromObject(newComp).ToYAML()
     
  }

}
