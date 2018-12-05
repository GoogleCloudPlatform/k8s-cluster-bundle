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
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/util/validation/field"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

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

// All validates components and components sets, providing as many errors as it can.
func All(cp []*bundle.ComponentPackage, cs *bundle.ComponentSet) field.ErrorList {
	errs := field.ErrorList{}
	errs = append(errs, ComponentSet(cs)...)
	errs = append(errs, AllComponents(cp)...)
	errs = append(errs, ComponentsAndComponentSet(cp, cs)...)
	errs = append(errs, AllComponentObjects(cp)...)
	return errs
}

func cPath(ref bundle.ComponentReference) *field.Path {
	return field.NewPath("Component").Key(fmt.Sprintf("%v", ref))
}

// AllComponents validates a list of components.
func AllComponents(components []*bundle.ComponentPackage) field.ErrorList {
	errs := field.ErrorList{}
	objCollect := make(map[bundle.ComponentReference]bool)
	for _, c := range components {
		errs = append(errs, Component(c)...)

		ref := c.ComponentReference()
		p := cPath(ref)
		if _, ok := objCollect[ref]; ok {
			errs = append(errs, field.Duplicate(p, fmt.Sprintf("component reference %v", ref)))
			continue
		}
		objCollect[ref] = true
	}
	return errs
}

// Component validates a single component.
func Component(c *bundle.ComponentPackage) field.ErrorList {
	errs := field.ErrorList{}
	pi := field.NewPath("Component")

	n := c.Spec.ComponentName
	if n == "" {
		errs = append(errs, field.Required(pi.Child("Spec", "ComponentName"), "components must have ComponentName"))
	}
	ver := c.Spec.Version
	if ver == "" {
		errs = append(errs, field.Required(pi.Child("Spec", "Version"), "components must have a Version"))
	}
	if n == "" || ver == "" {
		// Subsequent validation relies on components having a unique name+version pair.
		return errs
	}

	p := cPath(c.ComponentReference())

	if nameErrs := validateName(p.Child("Spec", "ComponentName"), n); len(errs) > 0 {
		errs = append(errs, nameErrs...)
	}

	api := c.APIVersion
	if !apiVersionPattern.MatchString(api) {
		errs = append(errs, field.Invalid(p.Child("APIVersion"), api, "must have the form \"bundle.gke.io/<version>\""))
	}

	expType := "ComponentPackage"
	if k := c.Kind; k != expType {
		errs = append(errs, field.Invalid(p.Child("Kind"), k, "kind must be ComponentPackage"))
	}

	if !versionPattern.MatchString(ver) {
		errs = append(errs, field.Invalid(p.Child("Spec", "Version"), ver, "must be of the form X.Y.Z"))
	}
	return errs
}

// ComponentSet validates a component.
func ComponentSet(cs *bundle.ComponentSet) field.ErrorList {
	p := field.NewPath("ComponentSet")

	errs := field.ErrorList{}
	if cs == nil {
		return errs
	}

	n := cs.Spec.SetName
	if n == "" {
		errs = append(errs, field.Required(p.Child("Spec", "SetName"), "setName is required"))
	}
	ver := cs.Spec.Version
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

	api := cs.APIVersion
	if !apiVersionPattern.MatchString(api) {
		errs = append(errs, field.Invalid(p.Child("APIVersion"), api, "must have an apiVersion of the form \"bundle.gke.io/<version>\""))
	}

	expType := "ComponentSet"
	if k := cs.Kind; k != expType {
		errs = append(errs, field.Invalid(p.Child("Kind"), k, "must be ComponentSet"))
	}

	if !versionPattern.MatchString(ver) {
		errs = append(errs, field.Invalid(p.Child("Spec", "Version"), ver, `must be of the form X.Y.Z`))
	}

	return errs
}

// ComponentsAndComponentSet validates components in the context of a component set.
func ComponentsAndComponentSet(components []*bundle.ComponentPackage, cs *bundle.ComponentSet) field.ErrorList {
	errs := field.ErrorList{}

	compMap := make(map[bundle.ComponentReference]*bundle.ComponentPackage)
	for _, c := range components {
		compMap[c.ComponentReference()] = c
	}

	p := field.NewPath("ComponentSet")
	// It is possible for there to be components that the component set does not
	// know about, but all components in the component set must be in the
	// components list
	for _, ref := range cs.Spec.Components {
		if _, ok := compMap[ref]; !ok {
			errs = append(errs, field.NotFound(p.Child("Spec", "Components").Key(fmt.Sprintf("%v", ref)), "component reference from set not found in component list"))
		}
	}

	return errs
}

// AllComponentObjects validates all objects in all componenst
func AllComponentObjects(components []*bundle.ComponentPackage) field.ErrorList {
	errs := field.ErrorList{}
	for _, ca := range components {
		errs = append(errs, ComponentObjects(ca)...)
	}
	return errs
}

// ComponentObjects validates objects in a componenst.
func ComponentObjects(cp *bundle.ComponentPackage) field.ErrorList {
	// Map to catch duplicate objects.
	compObjects := make(map[core.ObjectRef]bool)

	errs := field.ErrorList{}

	ref := cp.ComponentReference()
	basep := cPath(ref)
	for i, obj := range cp.Spec.Objects {
		n := obj.GetName()
		if n == "" {
			errs = append(errs, field.Required(basep.Child("Spec", "Objects").Index(i).Child("Metadata", "Name"),
				""))
			continue
		}
		p := basep.Child("Spec", "Objects").Key(n)

		if nameErrs := validateName(p, n); len(nameErrs) > 0 {
			errs = append(errs, nameErrs...)
		}

		oref := core.ObjectRefFromUnstructured(obj)
		if _, ok := compObjects[oref]; ok {
			errs = append(errs, field.Duplicate(p, fmt.Sprintf("object reference: %s", oref)))
		}
		compObjects[oref] = true
	}

	for i, objfile := range cp.Spec.ObjectFiles {
		url := objfile.URL
		fp := basep.Child("Spec", "ObjectFiles").Index(i)
		if urlErr := validateURL(fp, url); urlErr != nil {
			errs = append(errs, urlErr)
		}
	}
	return errs
}
