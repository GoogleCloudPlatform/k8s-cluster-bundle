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

package find

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

// ComponentFinder is a wrapper which allows for efficient searching through
// component data. The data is intended to be readonly; if modifications are
// made to the data, subsequent lookups will fail.
type ComponentFinder struct {
	nameCompLookup map[string][]*bundle.Component
	keyCompLookup  map[bundle.ComponentReference]*bundle.Component
	data           []*bundle.Component
}

// NewComponentFinder creates a new ComponentFinder or returns an error.
func NewComponentFinder(data []*bundle.Component) *ComponentFinder {
	nlup := make(map[string][]*bundle.Component)
	klup := make(map[bundle.ComponentReference]*bundle.Component)
	for _, comp := range data {
		name := comp.Spec.ComponentName
		klup[comp.ComponentReference()] = comp
		if list := nlup[name]; list == nil {
			nlup[name] = []*bundle.Component{comp}
		} else {
			nlup[name] = append(nlup[name], comp)
		}
	}
	return &ComponentFinder{
		nameCompLookup: nlup,
		keyCompLookup:  klup,
		data:           data,
	}
}

// Component returns the component package that matches a reference,
// returning nil if no match is found.
func (f *ComponentFinder) Component(ref bundle.ComponentReference) *bundle.Component {
	return f.keyCompLookup[ref]
}

// ComponentVersions returns the all the component versions for a given
// component name. The references are not sorted.
func (f *ComponentFinder) ComponentVersions(name string) []bundle.ComponentReference {
	comps := f.nameCompLookup[name]
	var refs []bundle.ComponentReference
	for _, c := range comps {
		refs = append(refs, c.ComponentReference())
	}
	return refs
}

// AllComponents return all the components known by the finder.
func (f *ComponentFinder) AllComponents() []*bundle.Component {
	var out []*bundle.Component
	for _, c := range f.keyCompLookup {
		out = append(out, c)
	}
	return out
}

// UniqueComponentFromName returns the single component package that matches a
// string-name. If no component is found, nil is returned. If there are two
// components that match the name, the method returns an error.
func (f *ComponentFinder) UniqueComponentFromName(name string) (*bundle.Component, error) {
	comps := f.ComponentVersions(name)
	if len(comps) == 0 {
		return nil, nil
	} else if len(comps) > 1 {
		return nil, fmt.Errorf("duplicate component found for name %q", name)
	}
	return f.Component(comps[0]), nil
}

// Objects returns Component's Cluster objects (given some object
// ref) or nil.
func (f *ComponentFinder) Objects(cref bundle.ComponentReference, ref core.ObjectRef) []*unstructured.Unstructured {
	comp := f.Component(cref)
	if comp == nil {
		return nil
	}
	return NewObjectFinder(comp).Objects(ref)
}

// ObjectsFromUniqueComponent gets the objects for a component, which
// has the same behavior as Objects, except that the component name is
// assumed to be unique (and so panics if that assumption does not hold).
func (f *ComponentFinder) ObjectsFromUniqueComponent(name string, ref core.ObjectRef) ([]*unstructured.Unstructured, error) {
	comp, err := f.UniqueComponentFromName(name)
	if err != nil {
		return nil, err
	}
	if comp == nil {
		return nil, nil
	}
	return NewObjectFinder(comp).Objects(ref), nil
}

// ObjectFinder finds objects within components
type ObjectFinder struct {
	component *bundle.Component
}

// NewObjectFinder returns an ObjectFinder instance.
func NewObjectFinder(component *bundle.Component) *ObjectFinder {
	return &ObjectFinder{component}
}

// Objects finds cluster objects matching a certain ObjectRef key. If
// the ObjectRef is partially filled out, then only those fields will be used
// for searching and the partial matches will be returned.
func (c *ObjectFinder) Objects(ref core.ObjectRef) []*unstructured.Unstructured {
	var out []*unstructured.Unstructured
	for _, o := range c.component.Spec.Objects {
		var key core.ObjectRef
		if ref.Name != "" {
			// Doing a search based on name
			key.Name = o.GetName()
		}
		if ref.APIVersion != "" {
			// Doing a search based on API version
			key.APIVersion = o.GetAPIVersion()
		}
		if ref.Kind != "" {
			// Doing a search based on kind
			key.Kind = o.GetKind()
		}
		if key == ref {
			out = append(out, o)
		}
	}
	return out
}
