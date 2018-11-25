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

package maker

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
)

// ObjHandler is a function that can 'make' objects.
type ObjHandler func(obj *unstructured.Unstructured, ref bundle.ComponentReference, pm ParamMaker) (*unstructured.Unstructured, error)

// MakeCommon provides common functionality for making components, deferring
// only the object handling logic.
func MakeCommon(comp *bundle.ComponentPackage, p ParamMaker, of *filter.Options, fn ObjHandler) (*bundle.ComponentPackage, error) {
	comp = comp.DeepCopy()
	ref := comp.ComponentReference()

	if len(comp.Spec.Objects) == 0 {
		return nil, fmt.Errorf("no objects found for component %v", ref)
	}

	// Filter the objects before handling them.
	objs := filter.Filter().Objects(comp.Spec.Objects, of)

	// Construct the objects.
	var newObj []*unstructured.Unstructured
	for _, obj := range objs {
		nob, err := fn(obj, ref, p)
		if err != nil {
			return nil, err
		}
		newObj = append(newObj, nob)
	}

	comp.Spec.Objects = newObj
	return comp, nil
}
