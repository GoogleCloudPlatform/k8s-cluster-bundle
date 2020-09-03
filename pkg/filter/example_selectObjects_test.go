package filter_test

import (
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Example_selectObjects() {
	bundle, err := converter.FromYAMLString(componentsYAMLs).ToBundle()
	if err != nil {
		log.Fatalf("could not convert from yaml to bundle: %v", err)
	}
	// If the match is on a component level then all objects are selected for that component.
	c1 := filter.SelectObjects(bundle.Components, filter.ComponentFieldMatchIn([]string{"zap", "zog"}, componentName))
	fmt.Println("c1:", toComponentList(c1))

	// Only the objects that match are returned on the component.
	c2 := filter.SelectObjects(bundle.Components, filter.Or(
		filter.ObjectFieldMatchIn([]string{"zap-pod"}, (*unstructured.Unstructured).GetName),
		filter.ComponentFieldMatchIn([]string{"nog"}, componentName)))
	fmt.Println("c2:", toComponentList(c2))

	// Only pods that are not in kube-system namespace are returned.
	c3 := filter.SelectObjects(bundle.Components, filter.And(
		filter.ObjectFieldMatchIn([]string{"Pod"}, (*unstructured.Unstructured).GetKind),
		filter.Not(filter.ObjectFieldMatchIn([]string{"kube-system"}, (*unstructured.Unstructured).GetNamespace))))
	fmt.Println("c3:", toComponentList(c3))
	// Output:
	// c1: [{zap [zap-pod zip-pod]} {zog [zog-pod zog-dep]}]
	// c2: [{zap [zap-pod]} {nog [nog-pod]}]
	// c3: [{zap [zip-pod]} {bog [bog-pod-2]} {zog [zog-pod]}]
}
