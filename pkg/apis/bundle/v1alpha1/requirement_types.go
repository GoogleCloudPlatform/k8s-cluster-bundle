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
// set of components, the Requirements are checked to ensure that the set
// of components is valid. Additionally, the Requirements may be used to
// construct the set of copmonents by selecting the appropriate components
// from a universe of available components.
//
// The structure and terminology is based on Go Modules. For a detailed
// discussion of Go Modules see: https://github.com/golang/go/wiki/Module.
type Requirements struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Require specific component versions. A component that with an AppVersion
	// >= to the AppVersion specified by a component may satisfy this
	// requirement, using minimal version selection or a similar algorithm.
	Require []ComponentRequire `json:"require,omitempty"`
}

// ComponentRequire is a specifies a minimal component version that will work
// with the component.
type ComponentRequire struct {
	// ComponentName (required) specifies the name of a component.
	ComponentName string `json:"componentName,omitempty"`

	// AppVersion specifies the minimum required AppVersion present in another
	// components Requirements object. In otherwords, the AppVersion given by
	// ComponentName in a Bundle or Set must be >= to this AppVersion, fixing the
	// major version.
	//
	// If AppVersion is not included, then any component with the specified
	// ComponentName will be considered a valid match.
	AppVersion string `json:"appVersion,omitempty"`
}
