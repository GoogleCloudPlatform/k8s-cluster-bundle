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

package core

import (
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// ComponentData encapsulates external component data. It can be used for
// build systems and libraries, but is not intended to be applied to Kubernetes
// clusters.
type ComponentData struct {
	// ComponentFiles reference component files.
	ComponentFiles []bundle.File `json:"componentFiles,omitempty"`

	// Components included inline in the component-data.
	Components []*bundle.ComponentPackage `json:"components,omitempty"`
}

// DeepCopy makes a deep copy of the ComponentData.
func (c *ComponentData) DeepCopy() *ComponentData {
	newdata := &ComponentData{}
	for _, c := range c.Components {
		newdata.Components = append(newdata.Components, c.DeepCopy())
	}
	return newdata

	for _, c := range c.ComponentFiles {
		// since files are not values, we can just assign to copy
		newdata.ComponentFiles = append(newdata.ComponentFiles, c)
	}
	return newdata
}
