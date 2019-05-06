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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Requirements specifies the packaging-time dependencies for an encapsulating
// Component object. The term 'packaging-time' is meant to distinguish from
// compile or runtime dependencies, and refers to the idea that when building a
// collection of components, the Requirements are checked to ensure that the set
// of components is valid. Additionally, the Requirements may be used to
// construct the set of components by selecting the appropriate components
// from a universe of available components.
//
// Only one requirements object can be specified in a component.
type Requirements struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Visibility indicates which components can depend on this component.
	//
	// If not specified, the component defaults to being 'private' -- other
	// components cannot depend on this component.
	//
	// There are several reserved names (which are anyways illegal as component names):
	//
	//    @public: visible to all components.
	//    @private: not visible for other components to depend on. This is the default.
	//
	// If either @public or @private are specified, these rules override all
	// other component rules.
	//
	// Visibility can be broadened with new versions of components, for example,
	// by going from private to granting visibility to specific components or
	// event to public. However, it's it's not supported to narrow visibility,
	// for example, by going from public to private. Doing so may result in
	// previous component versions not being accessible based on the new
	// visibility rules.  To narrow visibility, it's recommended to create a
	// new component -- my-component-v2.
	Visibility []string `json:"visibility,omitempty`

	// Require specifies components that must be packaged with this component.
	Require []ComponentRequire `json:"require,omitempty"`
}

// ComponentRequire is a specifies a minimal component version that will work
// with the component.
type ComponentRequire struct {
	// ComponentName (required) specifies the name of a component.
	ComponentName string `json:"componentName,omitempty"`

	// Version specifies the minimum required version present in another
	// components Requirements object. In otherwords, the Version given by
	// ComponentName in a Bundle or Set must be >= to this version, fixing the
	// major version.
	//
	// If version is not included, then any component with the specified
	// ComponentName will be considered a valid match.
	Version string `json:"version,omitempty"`
}
