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

package validate

import (
  "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
  "testing"
)

// COMPONENT TESTS
func assertValidateComponent(componentString string, numErrors int, t *testing.T) {
  component, errParse := converter.FromYAMLString(componentString).ToComponent()
  if errParse != nil {
    t.Fatalf("parse error: %v", errParse)
  }
  errs := Component(component)
  numErrorsActual := len(errs)
  if numErrorsActual != numErrors {
    t.Fatalf("expected %v errors got %v,  errors: %v", numErrors, numErrorsActual, errs)
  }
}
func TestComponent(t *testing.T) {
  component := `
    apiVersion: bundle.gke.io/v1alpha1
    kind: Component
    metadata:
      creationTimestamp: null
    spec:
      version: 30.0.2
      componentName: etcd-component`

  assertValidateComponent(component, 0, t)

}
func TestComponentNoRefs(t *testing.T) {
  component := `
    apiVersion: 'bundle.gke.io/v1alpha1'
    kind: Component
    metadata:
      name: foo-comp-1.0.2
    spec:
      componentName: foo-comp
      version: 2.10.1
      appVersion: 3.10.1`

  assertValidateComponent(component, 0, t)
}
func TestComponentVersionDotNotationAppVersion(t *testing.T) {
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

func TestComponentRequiresKind(t *testing.T) {
  component := `     # no kind field
  apiVersion: 'bundle.gke.io/v1alpha1'
  metadata:
    name: foo-comp-1.0.2
  spec:
    componentName: foo-comp
    version: 1.0.2`

  assertValidateComponent(component, 1, t)
}
func TestComponentXYZAppVersion(t *testing.T) {
  component := `
    apiVersion: 'bundle.gke.io/v1alpha1'
    kind: Component
    metadata:
      name: foo-comp-1.0.2
    spec:
      componentName: foo-comp
      version: 2.10.1
      appVersion: 3.10.32-blah.0` // invalid

  assertValidateComponent(component, 0, t)
}
func TestComponentSetInvalidVersion3(t *testing.T) {
  component := `
    apiVersion: 'bundle.gke.io/v1alpha1'       
    kind: Component
    metadata:
      name: foo-comp-1.0.2
    spec:
      componentName: foo-comp
      version: 2.010.1` // invalid

  assertValidateComponent(component, 1, t)
}
func TestComponentAppVersion(t *testing.T) {
  component := `
    apiVersion: 'bundle.gke.io/v1alpha1'
    kind: Component
    metadata:
      name: foo-comp-1.0.2
    spec:
      componentName: foo-comp
      version: 2.10.1
      appVersion: 2.010.1` // must be of the form X.Y.Z

  assertValidateComponent(component, 1, t)
}
func TestValidComponent(t *testing.T) {
  component := `
    apiVersion: 'bundle.gke.io/v1alpha1'
    kind: Component
    metadata:
      name: foo-comp-1.0.2
    spec:
      componentName: foo-comp
      version: 1.0.2
      objects:
      - apiVersion: v1
        kind: Pod
        metadata:
          name: foo-pod`

  assertValidateComponent(component, 0, t)
}
func TestComponentVersionMissing(t *testing.T) {
  component := `
    apiVersion: 'bundle.gke.io/v1alpha1'
    kind: Component
    metadata:
      name: foo-comp-1.0.2
    spec:
      componentName: foo-comp` // missing version: X.Y.Z in spec

  assertValidateComponent(component, 1, t)
}
func TestComponentMissingName(t *testing.T) {
  component := `
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
}

// Component Set Tests
func assertValidateComponentSet(componentSetString string, numErrors int, t *testing.T) {
  componentSet, errParse := converter.FromYAMLString(componentSetString).ToComponentSet()
  if errParse != nil {
    t.Fatalf("parse error: %v", errParse)
  }
  errs := ComponentSet(componentSet)
  numErrorsActual := len(errs)
  if numErrorsActual != numErrors {
    t.Fatalf("expected %v errors got %v:  %v", numErrors, numErrorsActual, errs)
  }
}

func TestComponentSet(t *testing.T) {
  defaultComponentSet := `
    apiVersion: 'bundle.gke.io/v1alpha1'
    kind: ComponentSet
    spec:
      setName: foo-set
      version: 1.0.2
      components:
      - componentName: foo-comp
        version: 1.0.2
        `
  assertValidateComponentSet(defaultComponentSet, 0, t)
}
func TestComponentSetInvalidVersion(t *testing.T) {
  invalidApiVersionComponentSet := `
    apiVersion: 'zork.gke.io/v1alpha1'         # this should be bundle.gke.io/<version>
    kind: ComponentSet
    spec:
      setName: zip
      version: 1.0.2
      components:
      - componentName: foo-comp
        version: 1.0.2`

  assertValidateComponentSet(invalidApiVersionComponentSet, 1, t)
}
func TestComponentSetInvalidVersion2(t *testing.T) {
  invalidVersionComponentSet := `
    apiVersion: 'bundle.gke.io/v1alpha1'
    kind: ComponentSet
    spec:
      setName: zip
      version: foo              # invalid version string, must be of form X.Y.Z
      components:
      - componentName: foo-comp
        version: 1.0.2`

  assertValidateComponentSet(invalidVersionComponentSet, 1, t)
}
func TestComponentSetMissingField(t *testing.T) {
  missingFieldComponentSet := `
    apiVersion: 'bundle.gke.io/v1alpha1'  # missing setName in spec
    kind: ComponentSet
    spec:
      version: 1.0.2
      components:
      - componentName: foo-comp
        version: 1.0.2`

  assertValidateComponentSet(missingFieldComponentSet, 1, t)
}
func TestComponentSetInvalidKind(t *testing.T) {
  invalidKindComponentSet := `
    apiVersion: 'bundle.gke.io/v1alpha1'    # kind Zor is not ComponentSet
    kind: Zor
    spec:
      setName: zip
      version: 1.0.2
      components:
      - componentName: foo-comp
        version: 1.0.2`

  assertValidateComponentSet(invalidKindComponentSet, 1, t)
}
func TestComponentSetMultipleErrors(t *testing.T) {
  componentSet := `      # missing apiVersion, kind
    spec:  
      setName: zip
      version: 1.0.2
      components:
      - componentName: zap-comp
        version: 1.0.2`

  assertValidateComponentSet(componentSet, 2, t)
}
