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
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/util/validation/field"

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
	// TODO(kashomon): Use the K8S version validation for this.
	// (k8s.io/apimachinery/pkg/util/version) after K8S v1.11
	versionPattern = regexp.MustCompile(fmt.Sprintf(`^%s\.%s\.%s$`, number, number, number))
)

// NewComponentValidator creates a new component Validator. The set of
// component packages is required, but if the component set is nil.
func NewComponentValidator(c []*bundle.ComponentPackage, set *bundle.ComponentSet) *ComponentValidator {
	return &ComponentValidator{c, set}
}

// Validate validates components and components sets, providing as many errors as it can.
func (v *ComponentValidator) Validate() field.ErrorList {
	errs := field.ErrorList{}
	errs = append(errs, v.validateComponentsSet()...)
	errs = append(errs, v.validateComponentPackages()...)
	errs = append(errs, v.validateObjects()...)
	return errs
}

func cPath(ref bundle.ComponentReference) *field.Path {
	return field.NewPath(fmt.Sprintf("Component{%s, %s}", ref.ComponentName, ref.Version))
}

func (v *ComponentValidator) validateComponentPackages() field.ErrorList {
	cpathIdx := func(idx int) *field.Path {
		return field.NewPath("Component").Index(idx)
	}

	errs := field.ErrorList{}
	objCollect := make(map[bundle.ComponentReference]bool)
	for i, ca := range v.components {
		n := ca.Spec.ComponentName
		if n == "" {
			errs = append(errs, field.Required(cpathIdx(i).Child("Spec", "ComponentName"), "componentName is required"))
		}
		ver := ca.Spec.Version
		if ver == "" {
			errs = append(errs, field.Required(cpathIdx(i).Child("Spec", "Version"), "version is required"))
		}
		if n == "" || ver == "" {
			// Other validation relies on components having a unique name+version pair.
			continue
		}

		ref := ca.ComponentReference()
		p := cPath(ref)

		if nameErrs := validateName(p.Child("Spec", "ComponentName"), n); len(errs) > 0 {
			errs = append(errs, nameErrs...)
		}

		api := ca.APIVersion
		if !apiVersionPattern.MatchString(api) {
			errs = append(errs, field.Invalid(p.Child("APIVersion").Index(i), api, "must have the form \"bundle.gke.io/<version>\""))
		}

		expType := "ComponentPackage"
		if k := ca.Kind; k != expType {
			errs = append(errs, field.Invalid(p.Child("Kind"), k, "kind must be ComponentPackage"))
		}

		if !versionPattern.MatchString(ver) {
			errs = append(errs, field.Invalid(p.Child("Spec", "Version"), ver, "must be of the form X.Y.Z"))
		}

		key := bundle.ComponentReference{ComponentName: n, Version: ver}
		if _, ok := objCollect[key]; ok {
			errs = append(errs, field.Duplicate(p, fmt.Sprintf("component key %v", key)))
			continue
		}
		objCollect[key] = true
	}
	return errs
}

func (v *ComponentValidator) validateComponentsSet() field.ErrorList {
	p := field.NewPath("ComponentSet")

	errs := field.ErrorList{}
	if v.componentSet == nil {
		return errs
	}

	n := v.componentSet.Spec.SetName
	if n == "" {
		errs = append(errs, field.Required(p.Child("Spec", "SetName"), "setName is required"))
	}
	ver := v.componentSet.Spec.Version
	if ver == "" {
		errs = append(errs, field.Required(p.Child("Spec", "Version"), "version is required"))
	}

	if n == "" || ver == "" {
		// Other validation relies on components having a unique name+version pair.
		return errs
	}

	if nameErrs := validateName(p.Child("Spec", "SetName"), n); len(nameErrs) > 0 {
		errs = append(errs, nameErrs...)
	}

	api := v.componentSet.APIVersion
	if !apiVersionPattern.MatchString(api) {
		errs = append(errs, field.Invalid(p.Child("APIVersion"), api, "must have an apiVersion of the form \"bundle.gke.io/<version>\""))
	}

	expType := "ComponentSet"
	if k := v.componentSet.Kind; k != expType {
		errs = append(errs, field.Invalid(p.Child("Kind"), k, "must be ComponentSet"))
	}

	if !versionPattern.MatchString(ver) {
		errs = append(errs, field.Invalid(p.Child("Spec", "Version"), ver, "must be of the form X.Y.Z"))
	}

	compMap := make(map[bundle.ComponentReference]*bundle.ComponentPackage)
	for _, c := range v.components {
		compMap[c.ComponentReference()] = c
	}

	// It is possible for there to be components that the component set does not
	// know about, but all components in the component set must be in the
	// components list
	for i, ref := range v.componentSet.Spec.Components {
		if _, ok := compMap[ref]; !ok {
			errs = append(errs, field.Duplicate(p.Child("Spec", "Components").Index(i), fmt.Sprintf("component ref %v", ref)))
		}
	}
	return errs
}

func (b *ComponentValidator) validateObjects() field.ErrorList {
	// Map to catch duplicate objects.
	compObjects := make(map[core.ObjectRef]bool)

	errs := field.ErrorList{}
	for _, ca := range b.components {
		ref := ca.ComponentReference()
		for i, obj := range ca.Spec.Objects {
			n := obj.GetName()
			if n == "" {
				errs = append(errs, field.Required(cPath(ref).Child("Spec", "Objects").Index(i), "metadata.name is required for objects"))
				continue
			}
			p := cPath(ref).Child("Spec", fmt.Sprintf("Objects[%s]", n))

			if nameErrs := validateName(p, n); len(nameErrs) > 0 {
				errs = append(errs, nameErrs...)
			}

			oref := core.ObjectRefFromUnstructured(obj)
			if _, ok := compObjects[oref]; ok {
				errs = append(errs, field.Duplicate(p, fmt.Sprintf("object reference: %s", oref)))
			}
			compObjects[oref] = true
		}

		for i, objfile := range ca.Spec.ObjectFiles {
			url := objfile.URL
			fp := cPath(ref).Child("Spec", fmt.Sprintf("ObjectFiles[%d]", i))
			if urlErr := validateURL(fp, url); urlErr != nil {
				errs = append(errs, urlErr)
			}
		}
	}
	return errs
}
