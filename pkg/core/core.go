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
	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// ClusterObjectKey is a key representing a specific cluster object.
type ClusterObjectKey struct {
	// ComponentName represents the name of the component the object lives within.
	ComponentName string

	// ObjectName represents the name of the object.
	ObjectName string
}

var EmptyClusterObjectKey = ClusterObjectKey{}

// ClusterObjectKeyFromProto creates a non-pointer ClusterObjectKey from a
// proto.
func ClusterObjectKeyFromProto(k *bpb.ClusterObjectKey) ClusterObjectKey {
	return ClusterObjectKey{
		ComponentName: k.GetComponentName(),
		ObjectName:    k.GetObjectName(),
	}
}

// ToProto creates a ClusterObjectKey proto from a ClusterObjectKey value.
func (k ClusterObjectKey) ToProto() *bpb.ClusterObjectKey {
	return &bpb.ClusterObjectKey{
		ComponentName: k.ComponentName,
		ObjectName:    k.ObjectName,
	}
}

// ObjectReference is a stripped-down version of the Kubernetes corev1.ObjectReference type.
type ObjectReference struct {
	// The API Version for an Object
	APIVersion string

	// The Kind for an Object
	Kind string

	// The Name of an Object
	Name string
}

// ClusterObjectKeyFromProto creates a non-pointer ClusterObjectKey from a
// proto.
func ObjectReferenceFromProto(k *bpb.ObjectReference) ObjectReference {
	return ObjectReference{
		APIVersion: k.GetApiVersion(),
		Kind:       k.GetKind(),
		Name:       k.GetName(),
	}
}

// ToProto creates a ObjectReference proto from a ObjectReference value.
func (k ObjectReference) ToProto() *bpb.ObjectReference {
	return &bpb.ObjectReference{
		ApiVersion: k.APIVersion,
		Kind:       k.Kind,
		Name:       k.Name,
	}
}
