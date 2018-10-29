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

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

// BundleValidator validates bundles.
type BundleValidator struct {
	Bundle *bpb.ClusterBundle
}

var (
	apiVersionPattern = regexp.MustCompile(`^gke.io/k8s-cluster-bundle/\w+$`)

	// Regex string for numbers without leading zeros.
	number = `([1-9]\d*|0)`
	// Regex string for dot separated alphanumeric identifier with hyphen.
	dotID  = `[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*`
	// The three regex strings below make up the SemVer 2.0.0 string.
	// MAJOR.MINOR.PATCH-preRelease+metaData
	majorMinorPatch = fmt.Sprintf(`%s\.%s\.%s`, number, number, number)
	preRelease      = fmt.Sprintf(`(-%v)?`, dotID)
	metaData        = fmt.Sprintf(`(\+%v)?`, dotID)
	// Matches SemVer strings (https://semver.org/).
	semVerPattern = regexp.MustCompile(fmt.Sprintf(`^%s%s%s$`, majorMinorPatch, preRelease, metaData))
)

// NewBundleValidator creates a new Bundle Validator
func NewBundleValidator(b *bpb.ClusterBundle) *BundleValidator {
	return &BundleValidator{b}
}

// Validate validates Bundles, providing as many errors as it can.
func (b *BundleValidator) Validate() []error {
	var errs []error
	errs = append(errs, b.validateBundle()...)
	errs = append(errs, b.validateComponentPackageNames()...)
	errs = append(errs, b.validateClusterObjNames()...)
	return errs
}

func (b *BundleValidator) validateBundle() []error {
	var errs []error
	n := b.Bundle.GetMetadata().GetName()
	if n == "" {
		errs = append(errs, fmt.Errorf("bundle name was empty, but must always be present"))
	}
	api := b.Bundle.GetApiVersion()
	if !apiVersionPattern.MatchString(api) {
		errs = append(errs, fmt.Errorf("bundle apiVersion must have form \"gke.io/k8s-cluster-bundle/<version>\". was %q", api))
	}
	k := b.Bundle.GetKind()
	if k != "ClusterBundle" {
		errs = append(errs, fmt.Errorf("bundle kind must be \"ClusterBundle\". was %q", k))
	}
	v := b.Bundle.GetSpec().GetVersion()
	if !semVerPattern.MatchString(v) {
		errs = append(errs, fmt.Errorf("cluster bundle spec version string is not a SemVer string, was '%v'", v))
	}
	return errs
}

func (b *BundleValidator) validateComponentPackageNames() []error {
	var errs []error
	objCollect := make(map[string]*bpb.ComponentPackage)
	for _, ca := range b.Bundle.GetSpec().GetComponents() {
		n := ca.GetMetadata().GetName()
		if n == "" {
			errs = append(errs, fmt.Errorf("cluster components must always have a name. was empty for config %v", ca))
			continue
		}
		api := ca.GetApiVersion()
		if !apiVersionPattern.MatchString(api) {
			errs = append(errs, fmt.Errorf("cluster components apiversion have the apiVersion of \"gke.io/k8s-cluster-bundle/<version>\". was %q for config %v", api, ca))
		}
		k := ca.GetKind()
		if k != "ComponentPackage" {
			errs = append(errs, fmt.Errorf("cluster component kind must be \"ComponentPackage\". was %q for config %v", k, ca))
		}
		v := ca.GetSpec().GetVersion()
		if !semVerPattern.MatchString(v) {
			errs = append(errs, fmt.Errorf("cluster component spec version is not a SemVer string, was '%v' for config %v", v, ca))
		}
		if _, ok := objCollect[n]; ok {
			errs = append(errs, fmt.Errorf("duplicate cluster component key %q when processing config %v", n, ca))
			continue
		}
		objCollect[n] = ca
	}

	// Validate the min requirements on each component.
	for _, ca := range b.Bundle.GetSpec().GetComponents() {
		for _, mr := range ca.GetSpec().GetRequirements() {
			if !semVerPattern.MatchString(mr.GetComponentApiVersion()) {
				errs = append(errs, fmt.Errorf("min requirement has invalid SemVer string, was '%v', from config: %v", mr.GetComponentApiVersion(), mr))
			}

			if _, ok := objCollect[mr.GetComponent()]; !ok {
				errs = append(errs, fmt.Errorf("required component %v does not exist, requirement of config: %v", mr.GetComponent(),  ca))
			}
		}
	}
	return errs
}

func (b *BundleValidator) validateClusterObjNames() []error {
	var errs []error
	// Map to catch duplicate objects.
	compObjects := make(map[core.ObjectRef]bool)
	for _, ca := range b.Bundle.GetSpec().GetComponents() {
		compName := ca.GetMetadata().GetName()
		for _, obj := range ca.GetSpec().GetClusterObjects() {
			// We could check if the GVK/ObjectRef is unique. But name can appear
			// multiple times.
			n := core.ObjectName(obj)
			if n == "" {
				errs = append(errs, fmt.Errorf("cluster components must always have a metadata.name. was empty for component %q", compName))
				continue
			}

			k := core.ObjectKind(obj)
			if k == "" {
				errs = append(errs, fmt.Errorf("cluster components must always have a kind. was empty for object %q in component %q", n, compName))
				continue
			}

			a := core.ObjectAPIVersion(obj)
			if a == "" {
				errs = append(errs, fmt.Errorf("cluster components must always have an API Version. was empty for object %q in component %q", n, compName))
				continue
			}

			ref := core.ObjectRefFromStruct(obj)
			if _, ok := compObjects[ref]; ok {
				errs = append(errs, fmt.Errorf("duplicate cluster object found with object reference %v for component %q", ref, compName))
			}
			compObjects[ref] = true
		}
	}
	return errs
}
