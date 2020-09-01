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

// Package filter contains methods for selecting and filtering lists of
// components and objects.
package filter

import (
	"strings"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Filter filters the components and objects to produce a new set of components.
type Filter struct{}

// NewFilter creates a new Filter.
func NewFilter() *Filter {
	return &Filter{}
}

// Options for filtering bundles. By default, if any of the options match, then
// the relevant component or object is removed. If InvertMatch is set, then the
// objects are kept instead of removed.
type Options struct {
	// Kinds represent the Kinds to filter on. Can either be unqualified ("Deployment")
	// or qualified ("apps/v1beta1,Pod"). Qualified kinds are often called
	// GroupVersionKind in the Kubernetes Schema.
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

	// InvertMatch indicates wether to return the opposite match.
	InvertMatch bool
}

// OptionsFromObjectSelector creates an Options Object from a
func OptionsFromObjectSelector(sel *bundle.ObjectSelector) *Options {
	// TODO(kashomon): Should probably make a copy.
	if sel == nil {
		return nil
	}

	opts := &Options{
		Kinds:       sel.Kinds,
		Names:       sel.Names,
		Annotations: sel.Annotations,
		Labels:      sel.Labels,
		Namespaces:  sel.Namespaces,
	}
	if sel.InvertMatch != nil {
		opts.InvertMatch = *sel.InvertMatch
	}
	return opts
}

// FilterComponents removes components based on the ObjectMeta properties of
// the components, returning a new cluster bundle with just filtered
// components. Filtering for components doesn't take into account the
// properties of the object-children of the components. This is the opposite
// matching from SelectComponents.
func (f *Filter) FilterComponents(data []*bundle.Component, o *Options) []*bundle.Component {
	matched, notMatched := f.PartitionComponents(data, o)
	if o.InvertMatch {
		return matched
	}
	return notMatched
}

// SelectComponents picks components based on the ObjectMeta properties of the
// components, returning a new cluster bundle with just filtered components.
// Filtering for components doesn't take into account the properties of the
// object-children of the components. This performs the opposte matching from
// FilterComponents.
func (f *Filter) SelectComponents(data []*bundle.Component, o *Options) []*bundle.Component {
	matched, notMatched := f.PartitionComponents(data, o)
	if o.InvertMatch {
		return notMatched
	}
	return matched
}

// PartitionComponents splits the components into matched and not matched sets.
// PartitionComponents ignores the InvertMatch option, since both matched and
// unmatched objects are returned. Thus, the options to partition are always
// treated as options for matching objects.
func (f *Filter) PartitionComponents(data []*bundle.Component, o *Options) ([]*bundle.Component, []*bundle.Component) {
	var newData []*bundle.Component
	for _, cp := range data {
		newData = append(newData, cp.DeepCopy())
	}
	data = newData

	// nil options should not imply any change.
	if o == nil {
		return data, nil
	}

	var matched []*bundle.Component
	var notMatched []*bundle.Component
	for _, c := range data {
		if MatchesComponent(c, o) {
			matched = append(matched, c)
		} else {
			notMatched = append(notMatched, c)
		}
	}
	return matched, notMatched
}

// FilterObjects removes objects based on the ObjectMeta properties of the objects,
// returning a new list with just filtered objects. This performs the opposite match from SelectObjects.
func (f *Filter) FilterObjects(data []*unstructured.Unstructured, o *Options) []*unstructured.Unstructured {
	matched, notMatched := f.PartitionObjects(data, o)
	if o.InvertMatch {
		return matched
	}
	return notMatched
}

// SelectObjects picks objects based on the ObjectMeta properties of the objects,
// returning a new list with just filtered objects.
func (f *Filter) SelectObjects(data []*unstructured.Unstructured, o *Options) []*unstructured.Unstructured {
	matched, notMatched := f.PartitionObjects(data, o)
	if o.InvertMatch {
		return notMatched
	}
	return matched
}

// PartitionObjects splits the objects into matched and not matched sets.
// PartitionObjects ignores the KeepOnly option, since both matched and
// unmatched objects are returned. Thus, the options to partition are always
// treated as options for matching objects.
func (f *Filter) PartitionObjects(data []*unstructured.Unstructured, o *Options) ([]*unstructured.Unstructured, []*unstructured.Unstructured) {
	var newData []*unstructured.Unstructured
	for _, oj := range data {
		newData = append(newData, oj.DeepCopy())
	}
	data = newData

	// nil options should not imply any change.
	if o == nil {
		return data, nil
	}

	var matched []*unstructured.Unstructured
	var notMatched []*unstructured.Unstructured
	for _, cp := range data {
		if MatchesObject(cp, o) {
			matched = append(matched, cp)
		} else {
			notMatched = append(notMatched, cp)
		}
	}
	return matched, notMatched
}

// objectData contains data about the object being filtered.
type objectData struct {
	// APIVersion of the object.
	apiVersion string

	// Kind of the object.
	kind string

	// name of the object. For unstructured objects, this is is metadata.name.
	// For components, this is ComponentName
	name string

	// meta is the ObjectMeta for an object.
	meta *metav1.ObjectMeta
}

// objectDataFromComponent returns ObjectData created from a component.
func objectDataFromComponent(c *bundle.Component) *objectData {
	return &objectData{
		apiVersion: c.APIVersion,
		kind:       c.Kind,
		name:       c.Spec.ComponentName,
		meta:       c.ObjectMeta.DeepCopy(),
	}
}

// newObjectData returns ObjectData created from an unstructured Object.
func newObjectData(uns *unstructured.Unstructured) *objectData {
	return &objectData{
		apiVersion: uns.GetAPIVersion(),
		kind:       uns.GetKind(),
		name:       uns.GetName(),
		meta:       converter.FromUnstructured(uns).ExtractObjectMeta(),
	}
}

// MatchesComponent returns true if the conditions match a Component.
func MatchesComponent(c *bundle.Component, o *Options) bool {
	return matches(objectDataFromComponent(c), o)
}

// MatchesObject returns true if the conditions match an object.
func MatchesObject(obj *unstructured.Unstructured, o *Options) bool {
	return matches(newObjectData(obj), o)
}

// Matches returns whether an object matches the given. The match functions
// does an AND of ORS. In otherwords:
//
// (name1 OR name2 OR name3) AND
// (kind2 OR kind2 OR kind3) AND etc.
func matches(d *objectData, o *Options) bool {
	if o == nil {
		return true
	}

	matchesKinds := true
	if len(o.Kinds) > 0 {
		matchesKinds = false
		for _, optk := range o.Kinds {
			objKind := d.kind
			if strings.ContainsRune(optk, ',') {
				// Assume this is a Qualified Kind match of the form
				// "apps/v1beta1,Deployment". Commas shouldn't be normally in a kind.
				objKind = d.apiVersion + "," + d.kind
			}
			if optk == objKind {
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

// ComponentPredicate is a func that returns true for components that match
// criteria and false otherwise.
type ComponentPredicate func(*bundle.Component) bool

// Select takes a list of components and a predicate and returns the
// components that match the predicate.
func Select(components []*bundle.Component, predicate ComponentPredicate) []*bundle.Component {
	var out []*bundle.Component
	for _, component := range components {
		if predicate(component) {
			out = append(out, component)
		}
	}
	return out
}

// UnstructuredKindIn checks if any of the components object.GetKind()
// are in kinds.
func UnstructuredKindIn(kinds ...string) ComponentPredicate {
	return func(component *bundle.Component) bool {
		for _, obj := range component.Spec.Objects {
			for _, kind := range kinds {
				if obj.GetKind() == kind {
					return true
				}
			}
		}
		return false
	}
}

// UnstructuredNamespaceIn checks if any of the components object.GetNamespace()
// are in namespaces.
func UnstructuredNamespaceIn(namespaces ...string) ComponentPredicate {
	return func(component *bundle.Component) bool {
		for _, obj := range component.Spec.Objects {
			for _, namespace := range namespaces {
				if obj.GetNamespace() == namespace {
					return true
				}
			}
		}
		return false
	}
}

// ComponentNameIn returns true if Component.Spec.ComponentName is in names.
func ComponentNameIn(names ...string) ComponentPredicate {
	return func(component *bundle.Component) bool {
		for _, name := range names {
			if component.Spec.ComponentName == name {
				return true
			}
		}
		return false
	}
}

// UnstructuredNameIn returns true if any of the components object.GetName()
// are in names.
func UnstructuredNameIn(names ...string) ComponentPredicate {
	return func(component *bundle.Component) bool {
		for _, obj := range component.Spec.Objects {
			for _, name := range names {
				if obj.GetName() == name {
					return true
				}
			}
		}
		return false
	}
}

// anyKeyValueMatch takes two non nil maps and returns true if any key value
// pair in a matches a key value pair in b. If the value in b is "" then it
// always matches a matching key in a.
func anyKeyValueMatch(a, b map[string]string) bool {
	for ak, av := range a {
		for bk, bv := range b {
			if ak == bk && (av == bv || bv == "") {
				return true
			}
		}
	}
	return false
}

// ComponentAnnotationsIn returns true if any of the annotations on the
// component match the specified annotations. If the value for a key in
// annotations is "", then match is true if the key matches.
func ComponentAnnotationsIn(annotations map[string]string) ComponentPredicate {
	return func(component *bundle.Component) bool {
		annots := component.GetAnnotations()
		if annots == nil {
			return false
		}
		return anyKeyValueMatch(annots, annotations)
	}
}

// UnstructuredAnnotationsIn returns true if any of the components objects
// annotations match the specified annotations. If the value for a key in
// annotations is "", then match is true if the key matches.
func UnstructuredAnnotationsIn(annotations map[string]string) ComponentPredicate {
	return func(component *bundle.Component) bool {
		for _, obj := range component.Spec.Objects {
			annots := obj.GetAnnotations()
			if annots == nil {
				return false
			}
			if anyKeyValueMatch(annots, annotations) {
				return true
			}
		}
		return false
	}
}

// ComponentLabelsIn returns true if any of the labels on the
// component match the specified labels. If the value for a key in
// labels is "", then match is true if the key matches.
func ComponentLabelsIn(labels map[string]string) ComponentPredicate {
	return func(component *bundle.Component) bool {
		componentLabels := component.GetLabels()
		if componentLabels == nil {
			return false
		}
		return anyKeyValueMatch(componentLabels, labels)
	}
}

// UnstructuredLabelsIn return true if any of the components objects labels
// match the specified labels. If the value for a key in labels is "", then
// match is true if the key matches.
func UnstructuredLabelsIn(labels map[string]string) ComponentPredicate {
	return func(component *bundle.Component) bool {
		for _, obj := range component.Spec.Objects {
			objLabels := obj.GetLabels()
			if objLabels == nil {
				return false
			}
			if anyKeyValueMatch(objLabels, labels) {
				return true
			}
		}
		return false
	}
}

// And returns true if every ComponentPredicate function returns true and
// returns false otherwise.
func And(componentPredicates ...ComponentPredicate) ComponentPredicate {
	return func(component *bundle.Component) bool {
		for _, predicate := range componentPredicates {
			if !predicate(component) {
				return false
			}
		}
		return true
	}
}

// Or returns true if any ComponentPredicate function returns true and
// returns false otherwise.
func Or(componentPredicates ...ComponentPredicate) ComponentPredicate {
	return func(component *bundle.Component) bool {
		for _, predicate := range componentPredicates {
			if predicate(component) {
				return true
			}
		}
		return false
	}
}
