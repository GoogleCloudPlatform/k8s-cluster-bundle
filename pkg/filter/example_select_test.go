package filter_test

import (
	"fmt"
	"log"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var componentsYAMLs = `
components:
- spec:
    componentName: zap
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: zap-pod
        labels:
          component: zork
        annotations:
          foo: bar
        namespace: kube-system
    - apiVersion: v1
      kind: Pod
      metadata:
        name: zip-pod
        labels:
          component: zork
        annotations:
          foo: baz
        namespace: default
- spec:
    componentName: bog
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: bog-pod-1
        labels:
          component: bork
        annotations:
          foof: yar
        namespace: kube-system
    - apiVersion: v1
      kind: Pod
      metadata: 
        name: bog-pod-2
        namespace: default
- spec:
    componentName: nog
    objects:
    - apiVersion: v1beta1
      kind: Pod
      metadata:
        name: nog-pod
        labels:
          component: nork
        annotations:
          foof: narf
        namespace: kube
- spec:
    componentName: zog
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: zog-pod
        namespace: kube-system
    - apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: zog-dep
        labels:
          component: zork
        annotations:
          zoof: zarf
        namespace: zube
- spec:
    componentName: kog
    objects:
    - apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: kog-dep
        namespace: kube-system`

func componentName(c *bundle.Component) string {
	return c.Spec.ComponentName
}

type component struct {
	name    string
	objects []string
}

func toComponent(comp *bundle.Component) component {
	c := component{name: componentName(comp)}
	for _, obj := range comp.Spec.Objects {
		c.objects = append(c.objects, obj.GetName())
	}
	return c
}

func toComponentList(components []*bundle.Component) []component {
	out := make([]component, len(components))
	for i, c := range components {
		out[i] = toComponent(c)
	}
	return out
}

func Example_select() {
	bundle, err := converter.FromYAMLString(componentsYAMLs).ToBundle()
	if err != nil {
		log.Fatalf("could not convert from yaml to bundle: %v", err)
	}
	c1 := filter.Select(bundle.Components, filter.ComponentFieldMatchIn([]string{"zap", "zog"}, componentName))
	fmt.Println("c1:", toComponentList(c1))

	c2 := filter.Select(bundle.Components, filter.Or(
		filter.ObjectFieldMatchIn([]string{"bog-pod-1"}, (*unstructured.Unstructured).GetName),
		filter.ComponentFieldMatchIn([]string{"nog"}, componentName)))
	fmt.Println("c2:", toComponentList(c2))

	c3 := filter.Select(bundle.Components, filter.And(
		filter.ObjectFieldMatchIn([]string{"Pod"}, (*unstructured.Unstructured).GetKind),
		filter.ObjectFieldMatchIn([]string{"kube-system"}, (*unstructured.Unstructured).GetNamespace)))
	fmt.Println("c3:", toComponentList(c3))
	// Output:
	// c1: [{zap [zap-pod zip-pod]} {zog [zog-pod zog-dep]}]
	// c2: [{bog [bog-pod-1 bog-pod-2]} {nog [nog-pod]}]
	// c3: [{zap [zap-pod zip-pod]} {bog [bog-pod-1 bog-pod-2]} {zog [zog-pod zog-dep]}]
}
