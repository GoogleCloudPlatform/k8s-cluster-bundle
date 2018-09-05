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
	errs = append(errs, b.validateImageConfigs()...)
	errs = append(errs, b.validateClusterAppNames()...)
	errs = append(errs, b.validateClusterAppObjectNames()...)
	return errs
}

func (b *BundleValidator) validateImageConfigs() []error {
	var errs []error
	imageConfigs := make(map[string]*bpb.ImageConfig)
	for _, nc := range b.Bundle.GetSpec().GetImageConfigs() {
		n := nc.GetName()
		if n == "" {
			errs = append(errs, fmt.Errorf("image configs must always have a name. was empty for config %v", nc))
			continue
		}
		if _, ok := imageConfigs[n]; ok {
			errs = append(errs, fmt.Errorf("duplicate image config key %q found when processing config %v", n, nc))
			continue
		}
		imageConfigs[n] = nc
	}
	return errs
}

func (b *BundleValidator) validateClusterAppNames() []error {
	var errs []error
	appConfigs := make(map[string]*bpb.ClusterApplication)
	for _, ca := range b.Bundle.GetSpec().GetClusterApps() {
		n := ca.GetName()
		if n == "" {
			errs = append(errs, fmt.Errorf("cluster applications must always have a name. was empty for config %v", ca))
			continue
		}
		if _, ok := appConfigs[n]; ok {
			errs = append(errs, fmt.Errorf("duplicate cluster application key %q when processing config %v", n, ca))
			continue
		}
		appConfigs[n] = ca
	}
	return errs
}

// appObjKey is a key for an app+object, for use in maps.
type appObjKey struct {
	appName string
	objName string
}

func (b *BundleValidator) validateClusterAppObjectNames() []error {
	var errs []error
	appObjects := make(map[appObjKey]*bpb.ClusterObject)
	for _, ca := range b.Bundle.GetSpec().GetClusterApps() {
		appName := ca.GetName()
		appConfigs := make(map[string]*bpb.ClusterObject)
		for _, obj := range ca.GetClusterObjects() {
			n := obj.GetName()
			if n == "" {
				errs = append(errs, fmt.Errorf("cluster applications objects must always have a name. was empty for app %q", appName))
				continue
			}
			if _, ok := appConfigs[n]; ok {
				errs = append(errs, fmt.Errorf("duplicate cluster application object key %q when processing app %q", n, appName))
				continue
			}
			key := appObjKey{appName: appName, objName: n}
			if _, ok := appObjects[key]; ok {
				errs = append(errs, fmt.Errorf("combination of application name %q and object name %q was not unique", n, appName))
				continue
			}
			appConfigs[n] = obj
			appObjects[appObjKey{appName: appName, objName: n}] = obj
		}
	}
	return append(errs, b.validateClusterOptionsKeys(appObjects)...)
}

func (b *BundleValidator) validateClusterOptionsKeys(appObjects map[appObjKey]*bpb.ClusterObject) []error {
	var errs []error
	for _, key := range b.Bundle.GetSpec().GetOptionsDefaults() {
		appObjKey := appObjKey{key.GetAppName(), key.GetObjectName()}
		if _, ok := appObjects[appObjKey]; !ok {
			errs = append(errs, fmt.Errorf("options specified with application name %q and object name %q was not found", key.GetAppName(), key.GetObjectName()))
			continue
		}
	}
	return errs
}
