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

package converter

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// FromUnstructured creates an UnstructuredConverter.
func FromUnstructured(o *unstructured.Unstructured) *UnstructuredConverter {
	return &UnstructuredConverter{o}
}

// UnstructuredConverter is a converter between unstructured.Unstructured and
// other types.
type UnstructuredConverter struct {
	o *unstructured.Unstructured
}

// ExtractObjectMeta creates an ObjectMeta object from an Unstructured object.
func (c *UnstructuredConverter) ExtractObjectMeta() *metav1.ObjectMeta {
	metadata := &metav1.ObjectMeta{}
	metadata.Name = c.o.GetName()
	metadata.GenerateName = c.o.GetGenerateName()
	metadata.Namespace = c.o.GetNamespace()
	metadata.SelfLink = c.o.GetSelfLink()
	metadata.UID = c.o.GetUID()
	metadata.ResourceVersion = c.o.GetResourceVersion()
	metadata.Generation = c.o.GetGeneration()
	metadata.CreationTimestamp = c.o.GetCreationTimestamp()
	metadata.DeletionTimestamp = c.o.GetDeletionTimestamp()
	metadata.DeletionGracePeriodSeconds = c.o.GetDeletionGracePeriodSeconds()
	metadata.Labels = c.o.GetLabels()
	metadata.Annotations = c.o.GetAnnotations()
	metadata.OwnerReferences = c.o.GetOwnerReferences()
	metadata.Initializers = c.o.GetInitializers()
	metadata.Finalizers = c.o.GetFinalizers()
	metadata.ClusterName = c.o.GetClusterName()
	return metadata
}

// ToObject converts an Unstructured object to an arbitrary interface via JSON.
func (c *UnstructuredConverter) ToObject(obj interface{}) error {
	// Note: The apimachinery library has a custom converter method for
	// converting from Unstructured to an arbitrary object. It has some
	// non-hermetic configuration (environment variables), so is not used here.
	json, err := FromObject(c.o).ToJSON()
	if err != nil {
		return err
	}
	return FromJSON(json).ToObject(obj)
}
