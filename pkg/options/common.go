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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// ObjHandler is a function that can apply options to a Kubernetes object.
type ObjHandler func(obj *unstructured.Unstructured, ref bundle.ComponentReference, opts JSONOptions) (*unstructured.Unstructured, error)

// ApplyCommon provides common functionality for applying options, deferring
// the specific object handling logic.
func ApplyCommon(comp *bundle.ComponentPackage, opts JSONOptions, objFn ObjHandler) (*bundle.ComponentPackage, error) {
	comp = comp.DeepCopy()
	ref := comp.ComponentReference()

	// Construct the objects.
	var newObj []*unstructured.Unstructured
	for _, obj := range comp.Spec.Objects {
		nob, err := objFn(obj, ref, opts)
		if err != nil {
			return nil, err
		}
		newObj = append(newObj, nob)
	}

	comp.Spec.Objects = newObj
	return comp, nil
}
