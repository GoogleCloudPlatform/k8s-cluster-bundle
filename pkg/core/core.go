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
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// ClusterObjectKey is a key representing a specific cluster object.
type ClusterObjectKey struct {
	// ComponentName represents the name of the component the object lives within.
	ComponentName string

	// ObjectName represents the name of the object.
	ObjectName string
}

var EmptyClusterObjectKey = ClusterObjectKey{}

// ObjectReference is a stripped-down version of the Kubernetes corev1.ObjectReference type.
type ObjectReference struct {
	// The API Version for an Object.
	APIVersion string

	// The Kind for an Object.
	Kind string

	// The Name of an Object.
	Name string
}

// ObjectName gets the Object name from a cluster object.
func ObjectName(obj *structpb.Struct) string {
	meta := obj.GetFields()["metadata"]
	if meta == nil {
		return ""
	}
	metaval := meta.GetStructValue()
	name := metaval.GetFields()["name"]
	if name == nil {
		return ""
	}
	return name.GetStringValue()
}

// ObjectName gets the Object name from a cluster object.
func ObjectKind(obj *structpb.Struct) string {
	kind := obj.GetFields()["kind"]
	if kind == nil {
		return ""
	}
	return kind.GetStringValue()
}

// ObjectName gets the Object name from a cluster object.
func ObjectAPIVersion(obj *structpb.Struct) string {
	apiVersion := obj.GetFields()["apiVersion"]
	if apiVersion == nil {
		return ""
	}
	return apiVersion.GetStringValue()
}
