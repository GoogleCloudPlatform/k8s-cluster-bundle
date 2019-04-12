// Copyright 2019 Google LLC
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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
)

// CreateName creates a name string to be used for ObjectMeta.Name. It is used
// to create standarized names for Components.  It assumes that the inName and
// version fields already conform to naming requirements as discussed in:
// k8s.io/docs/concepts/overview/working-with-objects/names/
func CreateName(inName, version string) string {
	if inName == "" {
		return ""
	}
	if version == "" {
		return inName
	}
	return inName + "-" + version
}

// GetLocalObjectRef creates a LocalObjectReference from a ComponentReference.
func (c ComponentReference) GetLocalObjectRef() corev1.LocalObjectReference {
	return corev1.LocalObjectReference{Name: CreateName(c.ComponentName, c.Version)}
}

// MakeAndSetName constructs the name from the Component's ComponentName
// and Version and stores the result in metadata.name.
func (c *Component) MakeAndSetName() {
	c.ObjectMeta.Name = CreateName(c.Spec.ComponentName, c.Spec.Version)
	return
}

// ComponentReference creates a ComponentReference from a component.
func (c *Component) ComponentReference() ComponentReference {
	return ComponentReference{
		ComponentName: c.Spec.ComponentName,
		Version:       c.Spec.Version,
	}
}
