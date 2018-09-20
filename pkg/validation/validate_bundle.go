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
		n := nc.GetName()
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
		n := ca.GetName()
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

// compObjKey is a key for an component+object, for use in maps.
type compObjKey struct {
	compName string
	objName  string
}

func (b *BundleValidator) validateClusterObjNames() []error {
	var errs []error
	compObjects := make(map[compObjKey]*bpb.ClusterObject)
	for _, ca := range b.Bundle.GetSpec().GetComponents() {
		compName := ca.GetName()
		objCollect := make(map[string]*bpb.ClusterObject)
		for _, obj := range ca.GetClusterObjects() {
			n := obj.GetName()
			if n == "" {
				errs = append(errs, fmt.Errorf("cluster components must always have a name. was empty for component %q", compName))
				continue
			}
			if _, ok := objCollect[n]; ok {
				errs = append(errs, fmt.Errorf("duplicate cluster component object key %q when processing component %q", n, compName))
				continue
			}
			key := compObjKey{compName: compName, objName: n}
			if _, ok := compObjects[key]; ok {
				errs = append(errs, fmt.Errorf("combination of cluster component object name %q and object name %q was not unique", n, compName))
				continue
			}
			objCollect[n] = obj
			compObjects[compObjKey{compName: compName, objName: n}] = obj
		}
	}
	return append(errs, b.validateClusterOptionsKeys(compObjects)...)
}

func (b *BundleValidator) validateClusterOptionsKeys(compObjects map[compObjKey]*bpb.ClusterObject) []error {
	var errs []error
	for _, key := range b.Bundle.GetSpec().GetOptionsExamples() {
		compObjKey := compObjKey{key.GetComponentName(), key.GetObjectName()}
		if _, ok := compObjects[compObjKey]; !ok {
			errs = append(errs, fmt.Errorf("options specified with cluster component name %q and object name %q was not found", key.GetComponentName(), key.GetObjectName()))
			continue
		}
	}
	return errs
}
