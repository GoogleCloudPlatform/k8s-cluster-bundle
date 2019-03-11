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
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

func assertValidateComponent(componentString string, expectedError string, numErrors int, t *testing.T) {
	component, errParse := converter.FromYAMLString(componentString).ToComponent()
	if errParse != nil {
		t.Fatalf("parse error: %v", errParse)
	}
	errs := Component(component)
	numErrorsActual := len(errs)
	if numErrorsActual != numErrors {
		t.Fatalf("got %v errors, expected %v", numErrorsActual, numErrors)
	}

	if err := testutil.CheckErrorCases(errs.ToAggregate(), expectedError); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestValidateComponents(t *testing.T) {
	components := []struct {
		componentConfig string
		expectedErrors  int
		description     string
		errorDesc       string
	}{
		{
			componentConfig: `
        apiVersion: bundle.gke.io/v1alpha1
        kind: Component
        metadata:
          creationTimestamp: null
        spec:
          version: 30.0.2
          componentName: etcd-component`,
			expectedErrors: 0,
			description:    "basic component validation",
		},
		{
			componentConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'
        kind: Component
        metadata:
          name: foo-comp-1.0.2
        spec:
          componentName: foo-comp
          version: 2.10.1
          appVersion: 3.10.1`,
			expectedErrors: 0,
			description:    "basic component no refs",
		},
		{
			componentConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'
        kind: Component
        metadata:
          name: foo-comp-1.0.2
        spec:
          componentName: foo-comp
          version: 2.10.1
          appVersion: 3.10`,
			expectedErrors: 0,
			description:    "dot notation app version verification",
		},
		{
			componentConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'       
        kind: Component
        metadata:
          name: foo-comp-1.0.2
        spec:
          componentName: foo-comp
          version: 2.010.1             #invalid`,
			expectedErrors: 1,
			description:    "version notation is incorrect",
			errorDesc:      "must be of the form X.Y.Z",
		},
		{
			componentConfig: `     # no kind field
        apiVersion: 'bundle.gke.io/v1alpha1'
        metadata:
          name: foo-comp-1.0.2
        spec:
          componentName: foo-comp
          version: 1.0.2`,
			expectedErrors: 1,
			description:    "requires kind field to be specified",
			errorDesc:      "kind must be Component",
		},
		{
			componentConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'
        kind: Component
        metadata:
          name: foo-comp-1.0.2
        spec:
          componentName: foo-comp
          version: 2.10.1
          appVersion: 2.010.1   #invalid`,
			expectedErrors: 1,
			description:    "app version validation must be of form X.Y.Z",
			errorDesc:      "must be of the form X.Y.Z or X.Y",
		},
		{
			componentConfig: `
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
              name: foo-pod`,
			expectedErrors: 0,
			description:    "component with objects specified",
		},
		{
			componentConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'
        kind: Component
        metadata:
          name: foo-comp-1.0.2
        spec:
          componentName: foo-comp`,
			expectedErrors: 1,
			description:    "no version specified for component in spec",
			errorDesc:      "components must have a Version",
		},
	}

	for _, component := range components {
		t.Run(component.description, func(t *testing.T) {
			assertValidateComponent(component.componentConfig, component.errorDesc, component.expectedErrors, t)
		})
	}
}

func assertValidateComponentSet(componentSetString string, expectedError string, numErrors int, t *testing.T) {
	componentSet, errParse := converter.FromYAMLString(componentSetString).ToComponentSet()
	if errParse != nil {
		t.Fatalf("parse error: %v", errParse)
	}
	errs := ComponentSet(componentSet)
	numErrorsActual := len(errs)
	if numErrorsActual != numErrors {
		t.Fatalf("got %v errors, want %v", numErrorsActual, numErrors)
	}

	if err := testutil.CheckErrorCases(errs.ToAggregate(), expectedError); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestValidateComponentSets(t *testing.T) {
	componentSets := []struct {
		componentSetConfig string
		expectedErrors     int
		description        string
		errorDesc          string
	}{
		{
			componentSetConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'
        kind: ComponentSet
        spec:
          setName: foo-set
          version: 1.0.2
          components:
          - componentName: foo-comp
            version: 1.0.2`,
			expectedErrors: 0,
			description:    "basic component set validation",
		},
		{
			componentSetConfig: `
        apiVersion: 'zork.gke.io/v1alpha1'         # this should be bundle.gke.io/<version>
        kind: ComponentSet
        spec:
          setName: zip
          version: 1.0.2
          components:
          - componentName: foo-comp
            version: 1.0.2`,
			expectedErrors: 1,
			description:    "invalid api version",
			errorDesc:      "must have an apiVersion of the form \"bundle.gke.io/<version>",
		},
		{
			componentSetConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'
        kind: ComponentSet
        spec:
          setName: zip
          version: foo              # invalid version string, must be of form X.Y.Z
          components:
          - componentName: foo-comp
            version: 1.0.2`,
			expectedErrors: 1,
			description:    "invalid spec version must be of form X.Y.Z",
			errorDesc:      "must be of the form X.Y.Z",
		},
		{
			componentSetConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'  # missing setName in spec
        kind: ComponentSet
        spec:
          version: 1.0.2
          components:
          - componentName: foo-comp
            version: 1.0.2`,
			expectedErrors: 1,
			description:    "missing set name",
			errorDesc:      "setName is required",
		},
		{
			componentSetConfig: `
        apiVersion: 'bundle.gke.io/v1alpha1'    # kind Zor is not ComponentSet
        kind: Zor
        spec:
          setName: zip
          version: 1.0.2
          components:
          - componentName: foo-comp
            version: 1.0.2`,
			expectedErrors: 1,
			description:    "kind must be component set",
			errorDesc:      "must be ComponentSet",
		},
		{
			componentSetConfig: `      
        spec:  
          setName: zip
          version: 1.0.2
          components:
          - componentName: zap-comp
            version: 1.0.2`,
			expectedErrors: 2,
			description:    "api version, kind must be specified",
			errorDesc:      "must be ComponentSet",
		},
	}

	for _, componentSet := range componentSets {
		assertValidateComponentSet(componentSet.componentSetConfig, componentSet.errorDesc, componentSet.expectedErrors, t)
	}
}
