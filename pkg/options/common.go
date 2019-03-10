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

package options

import (
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ObjHandler is a function that can apply options to a Kubernetes object.
type ObjHandler func(obj *unstructured.Unstructured, ref bundle.ComponentReference, opts JSONOptions) (*unstructured.Unstructured, error)

// ApplyCommon provides common functionality for applying options, deferring
// the specific object handling logic. The objects will be modified in-place;
// the caller should copy them if needed.
func ApplyCommon(ref bundle.ComponentReference, objs []*unstructured.Unstructured, opts JSONOptions, objFn ObjHandler) ([]*unstructured.Unstructured, error) {
	var newObj []*unstructured.Unstructured
	for _, obj := range objs {
		nob, err := objFn(obj, ref, opts)
		if err != nil {
			return nil, err
		}
		newObj = append(newObj, nob)
	}

	return newObj, nil
}

// PartitionObjectTemplates returns all the ObjectTemplates that match a
// specified TemplateKind and all other objects that don't match, in that
// order.
func PartitionObjectTemplates(allObjects []*unstructured.Unstructured, templateKind string) ([]*unstructured.Unstructured, []*unstructured.Unstructured) {
	var match []*unstructured.Unstructured
	var notMatch []*unstructured.Unstructured
	for _, obj := range allObjects {
		if obj.GetKind() != "ObjectTemplate" {
			notMatch = append(notMatch, obj)
			continue
		}

		var templateType string
		objData := obj.Object["templateType"]
		templateType, isString := objData.(string)
		if !isString {
			notMatch = append(notMatch, obj)
			continue
		}

		if templateType == templateKind {
			match = append(match, obj)
		} else {
			notMatch = append(notMatch, obj)
		}
	}
	return match, notMatch
}
