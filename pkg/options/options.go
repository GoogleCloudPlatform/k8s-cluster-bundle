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

// Package options provides functionality for adding options to components.
//
// Options, in this context, are values that need to be added to components at
// runtime. Sometimes these are called 'last-mile' customizations. For example,
// you might want the Cluster IP or the Cluster Name to a specific value.
//
// The parent options package provides common functionality and types while the
// subdirectories provide specific option-applying instances.
package options

import (
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// TODO(kashomon): Should options be of JSON Type or just interface{}?

// JSONOptions is an instance of options, represented as a JSON object encoded
// as map[string]interface{}. See more at the go docs for `encoding/json`.
type JSONOptions map[string]interface{}

// Applier represents an object that can take options and apply them to components.
type Applier interface {
	// ApplyOptions applys options to some subset objects from the component. The
	// returned component should be copy of the original, with (perhaps) modifications made to the original.
	ApplyOptions(comp *bundle.Component, opts JSONOptions) (*bundle.Component, error)
}
