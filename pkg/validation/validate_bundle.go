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

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

// BundleValidator validates bundles.
type BundleValidator struct {
	Bundle *bpb.ClusterBundle
}

// NewBundleValidator creates a new Bundle Validator
func NewBundleValidator(b *bpb.ClusterBundle) *BundleValidator {
	return &BundleValidator{b}
}

// Validate validates Bundles, providing as many errors as it can.
func (b *BundleValidator) Validate() []error {
	var errs []error
	errs = append(errs, b.validateNodeConfigs()...)
	errs = append(errs, b.validateClusterComponentNames()...)
	errs = append(errs, b.validateClusterObjNames()...)
	return errs
}

func (b *BundleValidator) validateNodeConfigs() []error {
	var errs []error
	nodeConfigs := make(map[string]*bpb.NodeConfig)
	for _, nc := range b.Bundle.GetSpec().GetNodeConfigs() {
		n := nc.GetMetadata().GetName()
		if n == "" {
			errs = append(errs, fmt.Errorf("node configs must always have a name. was empty for config %v", nc))
			continue
		}
		if _, ok := nodeConfigs[n]; ok {
			errs = append(errs, fmt.Errorf("duplicate node config key %q found when processing config %v", n, nc))
			continue
		}
		nodeConfigs[n] = nc
	}
	return errs
}

func (b *BundleValidator) validateClusterComponentNames() []error {
	var errs []error
	objCollect := make(map[string]*bpb.ClusterComponent)
	for _, ca := range b.Bundle.GetSpec().GetComponents() {
		n := ca.GetMetadata().GetName()
		if n == "" {
			errs = append(errs, fmt.Errorf("cluster components must always have a name. was empty for config %v", ca))
			continue
		}
		if _, ok := objCollect[n]; ok {
			errs = append(errs, fmt.Errorf("duplicate cluster component key %q when processing config %v", n, ca))
			continue
		}
		objCollect[n] = ca
	}
	return errs
}

func (b *BundleValidator) validateClusterObjNames() []error {
	var errs []error
	// Map to catch duplicate objects.
	compObjects := make(map[core.ObjectReference]bool)
	for _, ca := range b.Bundle.GetSpec().GetComponents() {
		compName := ca.GetMetadata().GetName()
		for _, obj := range ca.GetClusterObjects() {
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
