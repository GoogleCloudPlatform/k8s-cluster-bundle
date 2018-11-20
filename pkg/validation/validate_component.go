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

package validation

import (
	"errors"
	"fmt"
	"regexp"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

// ComponentValidator validates a set of components
type ComponentValidator struct {
	components   []*bundle.ComponentPackage
	componentSet *bundle.ComponentSet
}

var (
	apiVersionPattern = regexp.MustCompile(`^bundle.gke.io/\w+$`)

	// Regex string for numbers without leading zeros.
	number = `([1-9]\d*|0)`

	// Matches version string X.Y.Z, where X, Y and Z are non-negative integers
	// without leading zeros.
	versionPattern = regexp.MustCompile(fmt.Sprintf(`^%s\.%s\.%s$`, number, number, number))
)

// NewComponentValidator creates a new component Validator. The set of
// component packages is required, but if the component set is nil.
func NewComponentValidator(c []*bundle.ComponentPackage, set *bundle.ComponentSet) *ComponentValidator {
	return &ComponentValidator{c, set}
}

// Validate validates components and components sets, providing as many errors as it can.
func (v *ComponentValidator) Validate() []error {
	var errs []error
	errs = append(errs, v.validateComponentsSet()...)
	errs = append(errs, v.validateComponentPackages()...)
	errs = append(errs, v.validateObjects()...)
	return errs
}

func (v *ComponentValidator) validateComponentPackages() []error {
	var errs []error
	objCollect := make(map[bundle.ComponentReference]bool)
	for _, ca := range v.components {
		n := ca.Spec.ComponentName
		if err := ValidateName(n); err != nil {
			errs = append(errs, fmt.Errorf("the component name %q was invalid config %v", n, ca))
		}

		api := ca.APIVersion
		if !apiVersionPattern.MatchString(api) {
			errs = append(errs, fmt.Errorf("components APIVersion must have an apiVersion of the form \"bundle.gke.io/<version>\". was %q for config named %q", api, n))
		}

		expType := "ComponentPackage"
		if k := ca.Kind; k != expType {
			errs = append(errs, fmt.Errorf("component kind must be %q. was %q for config named %q", expType, k, n))
		}

		ver := ca.Spec.Version
		if ver == "" {
			errs = append(errs, errors.New("component spec version missing"))
		} else if !versionPattern.MatchString(ver) {
			errs = append(errs, fmt.Errorf("component spec version is invalid. was %q for component %q but must be of the form X.Y.Z", ver, n))
		}

		key := bundle.ComponentReference{ComponentName: n, Version: ver}
		if _, ok := objCollect[key]; ok {
			errs = append(errs, fmt.Errorf("duplicate component key %v", key))
			continue
		}
		objCollect[key] = true
	}
	return errs
}

func (v *ComponentValidator) validateComponentsSet() []error {
	var errs []error
	if v.componentSet == nil {
		return nil
	}
	n := v.componentSet.Spec.SetName
	if err := ValidateName(n); err != nil {
		errs = append(errs, fmt.Errorf("the component set name %q was invalid config %v", n, v.componentSet))
	}

	api := v.componentSet.APIVersion
	if !apiVersionPattern.MatchString(api) {
		errs = append(errs, fmt.Errorf("component set APIVersion must have an apiVersion of the form \"bundle.gke.io/<version>\". was %q", api))
	}

	expType := "ComponentSet"
	if k := v.componentSet.Kind; k != expType {
		errs = append(errs, fmt.Errorf("component set kind must be %q. was %q", expType, k))
	}

	ver := v.componentSet.Spec.Version
	if ver == "" {
		errs = append(errs, errors.New("component set spec version missing"))
	} else if !versionPattern.MatchString(ver) {
		errs = append(errs, fmt.Errorf("component set spec version is invalid. was %q but must be of the form X.Y.Z", ver))
	}

	compMap := make(map[bundle.ComponentReference]*bundle.ComponentPackage)
	for _, c := range v.components {
		compMap[c.ComponentReference()] = c
	}

	// It is possible for there to be components that the component set does not
	// know about, but all components in the component set must be in the
	// components list
	for _, ref := range v.componentSet.Spec.Components {
		if _, ok := compMap[ref]; !ok {
			errs = append(errs, fmt.Errorf("could not find component reference %v for any of the components", ref))
		}
	}
	return errs
}

func (b *ComponentValidator) validateObjects() []error {
	var errs []error
	// Map to catch duplicate objects.
	compObjects := make(map[core.ObjectRef]bool)
	for _, ca := range b.components {
		compName := ca.Spec.ComponentName
		for i, obj := range ca.Spec.Objects {
			n := obj.GetName()
			if n == "" {
				errs = append(errs, fmt.Errorf("objects must always have a metadata.name. was empty object %d in component %q", i, compName))
				continue
			}
			if err := ValidateName(n); err != nil {
				errs = append(errs, fmt.Errorf("invalid name %q for object %d: %v", n, i, err))
			}

			ref := core.ObjectRefFromUnstructured(obj)
			if _, ok := compObjects[ref]; ok {
				errs = append(errs, fmt.Errorf("duplicate object found with object reference %v for component %q", ref, compName))
			}
			compObjects[ref] = true
		}
	}
	return errs
}
