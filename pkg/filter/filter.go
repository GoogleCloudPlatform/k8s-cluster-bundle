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

package filter

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
)

// Filter filters the components and objects to produce a new set of components.
type Filter struct {
	// ChangeInPlace controls whether to change the objects in place or to make
	// copies. By default, the Filter returns deep copies of objects.
	ChangeInPlace bool
}

// NewFilter creates a new Filter.
func NewFilter() *Filter {
	return &Filter{}
}

// Options for filtering bundles. By default, if any of the options match, then
// the relevant component or object is removed. If KeepOnly is set, then the
// objects are kept instead of removed.
type Options struct {
	// Kinds represent the Kinds to filter on.
	Kinds []string

	// Names represent the names to filter on. For objects, this is the
	// metadata.name field. For components, this is the ComponentName.
	Names []string

	// Annotations contain key/value pairs to filter on. An empty string value matches
	// all annotation-values for a particular key.
	Annotations map[string]string

	// Labels contain key/value pairs to filter on. An empty string value matches
	// all label-values for a particular key.
	Labels map[string]string

	// Namespaces to filter on.
	Namespaces []string

	// KeepOnly means that instead of removing the found objects.
	KeepOnly bool
}

// Components filters components based on the ObjectMeta properties of
// the components, returning a new cluster bundle with just filtered
// components. By default components are removed, unless KeepOnly is set, and
// then the opposite is true. Filtering for components doesn't take into
// account the properties of the object-children of the components.
func (f *Filter) Components(data []*bundle.Component, o *Options) []*bundle.Component {
	if !f.ChangeInPlace {
		var newData []*bundle.Component
		for _, cp := range data {
			newData = append(newData, cp.DeepCopy())
		}
		data = newData
	}
	// nil options should not imply any change.
	if o == nil {
		return data
	}

	var matched []*bundle.Component
	var notMatched []*bundle.Component
	for _, c := range data {
		od := &objectData{
			kind: c.Kind,
			name: c.Spec.ComponentName,
			meta: &c.ObjectMeta,
		}
		matches := filterMeta(od, o)
		if matches {
			matched = append(matched, c)
		} else {
			notMatched = append(notMatched, c)
		}
	}
	if o.KeepOnly {
		return matched
	}
	return notMatched
}

// Objects filters objects based on the ObjectMeta properties of
// the objects, returning a new list with just filtered
// objects. By default objects are removed, unless KeepOnly is set, and
// then the opposite is true.
func (f *Filter) Objects(data []*unstructured.Unstructured, o *Options) []*unstructured.Unstructured {
	matched, notMatched := f.PartitionObjects(data, o)

	if o == nil || o.KeepOnly {
		return matched
	}

	return notMatched
}

// PartitionObjects splits the objects into matched and not matched sets.
func (f *Filter) PartitionObjects(data []*unstructured.Unstructured, o *Options) ([]*unstructured.Unstructured, []*unstructured.Unstructured) {
	if !f.ChangeInPlace {
		var newData []*unstructured.Unstructured
		for _, oj := range data {
			newData = append(newData, oj.DeepCopy())
		}
		data = newData
	}
	// nil options should not imply any change.
	if o == nil {
		return data, data
	}

	var matched []*unstructured.Unstructured
	var notMatched []*unstructured.Unstructured
	for _, cp := range data {
		od := &objectData{
			kind: cp.GetKind(),
			name: cp.GetName(),
			meta: converter.FromUnstructured(cp).ExtractObjectMeta(),
		}
		matches := filterMeta(od, o)
		if matches {
			matched = append(matched, cp)
		} else {
			notMatched = append(notMatched, cp)
		}
	}
	return matched, notMatched
}

// objectData contains data about the object being filtered.
type objectData struct {
	// Kind of the object
	kind string

	// name of the object. For unstructured objects, this is is metadata.name.
	// For components, this is ComponentName
	name string

	meta *metav1.ObjectMeta
}

// filterMeta returns whether an object matches the given
func filterMeta(d *objectData, o *Options) bool {
	matchesKinds := true
	if len(o.Kinds) > 0 {
		matchesKinds = false
		for _, optk := range o.Kinds {
			if optk == d.kind {
				matchesKinds = true
				break
			}
		}
	}
	matchesNS := true
	if len(o.Namespaces) > 0 {
		matchesNS = false
		for _, optn := range o.Namespaces {
			if optn == d.meta.Namespace {
				matchesNS = true
				break
			}
		}
	}
	matchesNames := true
	if len(o.Names) > 0 {
		matchesNames = false
		for _, optn := range o.Names {
			if optn == d.meta.GetName() {
				matchesNames = true
				break
			}
		}
	}
	matchesAnnot := true
	if len(o.Annotations) > 0 {
		matchesAnnot = false
		for key, v := range o.Annotations {
			if val, ok := d.meta.Annotations[key]; ok && val == v {
				matchesAnnot = true
				break
			}
		}
	}
	matchesLabels := true
	if len(o.Labels) > 0 {
		matchesLabels = false
		for key, v := range o.Labels {
			if val, ok := d.meta.Labels[key]; ok && val == v {
				matchesLabels = true
				break
			}
		}
	}
	return matchesKinds && matchesNS && matchesNames && matchesAnnot && matchesLabels
}
