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

// Package maker provides functionality for adding options to components
package maker

import (
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
)

// JSONMap is represents a JSON object encoded as map[string]interface{} where
// the children can be either map[string]interface{}, []interface{} or
// primitive type.
type JSONMap map[string]interface{}

// TODO(kashomon): Should parameters be of JSON Type or just interface{}

// ParamMaker makes parameters for the
type ParamMaker func() (JSONMap, error)

// ComponentMaker represents an object that can take options and apply them to components.
type ComponentMaker interface {
	// MakeComponent applys parameters to some subset objects from the component.
	MakeComponent(comp *bundle.ComponentPackage, p ParamMaker, objectFilter *filter.Options) (*bundle.ComponentPackage, error)
}
