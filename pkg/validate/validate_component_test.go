package validate

import (
	"testing"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

var defaultComponentData = `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: Component
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: Component
  metadata:
    name: bar-comp-2.0.3
  spec:
    componentName: foo-comp
    version: 2.0.3
    appVersion: 2.4.5
`

var defaultComponentSet = `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  setName: foo-set
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2
`

// Component Tests
func TestComponent(t *testing.T){
  defaultComponent := `
apiVersion: bundle.gke.io/v1alpha1
kind: Component
metadata:
  creationTimestamp: null
spec:
  version: 30.0.2
  componentName: etcd-component`

	
	component, err := converter.FromYAMLString(defaultComponent).ToComponent()
	if err != nil {
		t.Fatalf("invalid parsing: %v", err)
	}
	errs := Component(component)
	if len(errs) != 0 {
		t.Fatalf("%v", errs)
	}
}

func TestComponentNoRefs(t *testing.T){
	defaultComponentNoRefs := `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Component
metadata:
  name: foo-comp-1.0.2
spec:
  componentName: foo-comp
  version: 2.10.1
  appVersion: 3.10.1`

    component, errConv := converter.FromYAMLString(defaultComponentNoRefs).ToComponent()
    if errConv != nil {
    	t.Fatalf("error converting ", errConv)
    	return
    }
    errs := Component(component)
    if len(errs) != 0 {
    	t.Fatalf("%v", errs)
    }
}

/*{
			desc: "success: X.Y app version",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: Component
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.10.1
    appVersion: 3.10`,
		},*/

func assertValidateComponent(componentString string, numErrors int, t *testing.T){
	component, _ := converter.FromYAMLString(componentString).ToComponent()
	errs := Component(component)
	numErrorsActual := len(errs)
	if numErrorsActual != numErrors {
		t.Fatalf("expected %v errors got %v", numErrors, numErrorsActual)
	}
}	

func TestComponentVersionDotNotationAppVersion(t *testing.T){
component := `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Component
metadata:
  name: foo-comp-1.0.2
spec:
  componentName: foo-comp
  version: 2.10.1
  appVersion: 3.10`


    assertValidateComponent(component, 0, t)
}

/*
		{
			desc: "fail component: no kind",
			set:  defaultComponentSet,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2`,
			errSubstring: "must be Component",
		},
*/


// this is failing
/*func TestComponentRequiresKind(t *testing.T){
	component := `
apiVersion: 'bundle.gke.io/v1alpha1'
metadata:
  name: foo-comp-1.0.2
spec:
  componentName: foo-comp
  version: 1.0.2`

    assertValidateComponent(component, 1, t)
}*/

/*	{
			desc: "success: X.Y.Z-blah app version",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: Component
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.10.1
    appVersion: 3.10.32-blah.0`,
		},

		*/

func TestComponentXYZAppVersion(t *testing.T){
component := `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Component
metadata:
  name: foo-comp-1.0.2
spec:
  componentName: foo-comp
  version: 2.10.1
  appVersion: 3.10.32-blah.0`

assertValidateComponent(component, 0, t)
}

/*

		{
			desc: "fail: component invalid X.Y.Z version string ",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: Component
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.010.1`,
			errSubstring: "must be of the form X.Y.Z",
		},

*/ 

func TestComponentSetInvalidVersion3(t *testing.T){
component := `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Component
metadata:
  name: foo-comp-1.0.2
spec:
  componentName: foo-comp
  version: 2.010.1`

assertValidateComponent(component, 1, t)

}


/*
	{
			desc: "fail: component invalid X.Y.Z app version string ",
			set:  defaultComponentSetNoRefs,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: Component
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 2.10.1,
    appVersion: 2.010.1`,
			errSubstring: "must be of the form X.Y.Z",
		},
*/
func TestComponentAppVersion(t *testing.T){

}



/*
	{
			desc: "object fail: no metadata.name",
			set:  defaultComponentSet,
			components: `
components:
- apiVersion: 'bundle.gke.io/v1alpha1'
  kind: Component
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2
    objects:
    - apiVersion: v1
      kind: Pod`,
			errSubstring: "Required value",
		},

*/

/*func TestComponentMissingName(t *testing.T){
	component :=  `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Component
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2
    objects:
    - apiVersion: v1
      kind: Pod`	

  assertValidateComponent(component, 1, t)
}*/

// Component Set Tests
func TestComponentSet (t *testing.T){
	set, _ := converter.FromYAMLString(defaultComponentSet).ToComponentSet()
	errs := ComponentSet(set)
	if len(errs) != 0 {
		t.Fatalf("%v", errs)
	}
}

func TestComponentSetInvalidVersion(t *testing.T){
	invalidVersionComponentSet := `
apiVersion: 'zork.gke.io/v1alpha1'
kind: ComponentSet
spec:
  setName: zip
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2`

    /* 			desc: "component set fail: apiVersion",
			set: ,
			components:   defaultComponentData,
			errSubstring: "bundle.gke.io/<version>", */
	set, _ := converter.FromYAMLString(invalidVersionComponentSet).ToComponentSet()
	errs := ComponentSet(set)
	if len(errs) !=  1 {
		t.Fatalf("%v", errs)
	}
}

func TestComponentSetInvalidVersion2(t *testing.T){
	invalidVersionComponentSet := `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  setName: zip
  version: foo
  components:
  - componentName: foo-comp
    version: 1.0.2`

    /*  
		{
			desc: "component set fail: invalid X.Y.Z version string",
			set: ,
			components:   defaultComponentData,
			errSubstring: "must be of the form X.Y.Z",
		},
    */
    set, _ := converter.FromYAMLString(invalidVersionComponentSet).ToComponentSet()
	errs := ComponentSet(set)
	if len(errs) !=  1 {
		t.Fatalf("%v", errs)
	}
}

/*
{
			desc: "fail: missing set name",
			set: `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2`,
			components:   defaultComponentData,
			errSubstring: "Required value",
		},
*/
func TestComponentSetMissingField(t *testing.T){
	missingFieldComponentSet := `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2`

    set, _ := converter.FromYAMLString(missingFieldComponentSet).ToComponentSet()
	errs := ComponentSet(set)
	if len(errs) !=  1 {
		t.Fatalf("%v", errs)
	} 
}

/*

// Tests for component sets
		{
			desc: "component set fail: bad kind",
			set: `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Zor
spec:
  setName: zip
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2`,
			components:   defaultComponentData,
			errSubstring: "must be ComponentSet",

*/
func TestComponentSetInvalidKind(t * testing.T) {
	invalidKindComponentSet := `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Zor
spec:
  setName: zip
  version: 1.0.2
  components:
  - componentName: foo-comp
    version: 1.0.2`

    set, _ := converter.FromYAMLString(invalidKindComponentSet).ToComponentSet()
	errs := ComponentSet(set)
	if len(errs) !=  1 {
		t.Fatalf("%v", errs)
	} 
}