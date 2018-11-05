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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// ComponentKey references a component and provides a unique identifier for the
// component.
type ComponentKey struct {
	// CanonicalName of the component
	CanonicalName string

	// Version of the component, as a sem-ver string.
	Version string
}

func KeyFromComponent(c *bundle.ComponentPackage) ComponentKey {
	return ComponentKey{
		CanonicalName: c.Spec.CanonicalName,
		Version:       c.Spec.Version,
	}
}

var EmptyComponentKey = ComponentKey{}

// ClusterObjectKey is a key representing a specific cluster object.
type ClusterObjectKey struct {
	// Component references a single component.
	Component ComponentKey

	// Object represents a unique key for this object within a component.
	Object ObjectRef
}

var EmptyClusterObjectKey = ClusterObjectKey{}

// TODO(kashomon): Replace ObjectRef with corev1.TypedLocalObjectReference

// ObjectRef is a stripped-down version of the Kubernetes corev1.ObjectReference type.
type ObjectRef struct {
	// The API Version for an Object.
	APIVersion string

	// The Kind for an Object.
	Kind string

	// The Name of an Object.
	Name string
}

func ObjectRefFromUnstructured(o *unstructured.Unstructured) ObjectRef {
	return ObjectRef{
		APIVersion: o.GetAPIVersion(),
		Kind:       o.GetKind(),
		Name:       o.GetName(),
	}
}
