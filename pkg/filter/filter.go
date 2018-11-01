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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

// Filterer filters the components and objects to produce a new set of components.
type Filterer struct {
	data []*bundle.ComponentPackage
}

func NewFilterer(comp []*bundle.ComponentPackage) *Filterer {
	return &Filterer{comp}
}

// Options for filtering bundles. By default, if any of the options match, then
// the relevant component or object is removed. If KeepOnly is set, then the
// objects are kept instead of removed.
type Options struct {
	// Kinds represent the Kinds to filter on.
	Kinds []string

	// TODO(kashomon): Support filtering on component names?

	// Names represent the metadata.names to filter on.
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

// FilterComponents filters components based on the ObjectMeta properties of
// the components, returning a new cluster bundle with just filtered
// components. By default components are removed, unless KeepOnly is set, and
// then the opposite is true. Filtering for components doesn't take into
// account the properties of the object-children of the components.
func (f *Filterer) FilterComponents(o *Options) []*bundle.ComponentPackage {
	data := (&core.ComponentData{Components: f.data}).DeepCopy().Components
	var matched []*bundle.ComponentPackage
	var notMatched []*bundle.ComponentPackage
	for _, c := range data {
		matches := filterMeta(c.Kind, &c.ObjectMeta, o)
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

// FilterObjects filters objects based on the ObjectMeta properties of
// the objects, returning a new cluster bundle with just filtered
// objects. By default objectsare removed, unless KeepOnly is set, and
// then the opposite is true.
func (f *Filterer) FilterObjects(o *Options) []*bundle.ComponentPackage {
	data := (&core.ComponentData{Components: f.data}).DeepCopy().Components
	for _, cp := range data {
		var matched []*unstructured.Unstructured
		var notMatched []*unstructured.Unstructured
		for _, c := range cp.Spec.Objects {
			matches := filterMeta(c.GetKind(), converter.FromUnstructured(c).ExtractObjectMeta(), o)
			if matches {
				matched = append(matched, c)
			} else {
				notMatched = append(notMatched, c)
			}
		}
		if o.KeepOnly {
			cp.Spec.Objects = matched
		} else {
			cp.Spec.Objects = notMatched
		}
	}
	return data
}

func filterMeta(kind string, meta *metav1.ObjectMeta, o *Options) bool {
	for _, k := range o.Kinds {
		if k == kind {
			return true
		}
	}
	for _, n := range o.Namespaces {
		if n == meta.Namespace {
			return true
		}
	}
	// We ignore kind, because Kind should always be cluster bundle.
	for _, n := range o.Names {
		if meta.GetName() == n {
			return true
		}
	}
	if len(meta.Annotations) > 0 {
		for key, v := range o.Annotations {
			if val, ok := meta.Annotations[key]; ok && val == v {
				return true
			}
		}
	}
	if len(meta.Labels) > 0 {
		for key, v := range o.Labels {
			if val, ok := meta.Labels[key]; ok && val == v {
				return true
			}
		}
	}
	return false
}
