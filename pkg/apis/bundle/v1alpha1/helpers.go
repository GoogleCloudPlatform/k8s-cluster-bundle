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

package v1alpha1

import (
	"fmt"
	"net/url"

	corev1 "k8s.io/api/core/v1"
)

// CreateName creates a name string to be used for ObjectMeta.Name. It
// is used to create standarized names for ComponentPackages and ComponentSets.
// It assumes that the inName and version fields already conform to naming
// requirements as discussed in:
// k8s.io/docs/concepts/overview/working-with-objects/names/
func CreateName(inName, version string) string {
	if inName == "" {
		return ""
	}
	if version == "" {
		return inName
	}
	return inName + "-" + version
}

// GetLocalObjectRef creates a LocalObjectReference from a ComponentReference.
func (c ComponentReference) GetLocalObjectRef() corev1.LocalObjectReference {
	return corev1.LocalObjectReference{Name: CreateName(c.ComponentName, c.Version)}
}

// GetAllLocalObjectRefs creates LocalObjectReferences for all the component
// references in a ComponentSet.
func (c *ComponentSet) GetAllLocalObjectRefs() []corev1.LocalObjectReference {
	var out []corev1.LocalObjectReference
	for _, cp := range c.Spec.Components {
		out = append(out, cp.GetLocalObjectRef())
	}
	return out
}

// MakeAndSetName constructs the name from the ComponentSet's SetName and
// Version and stores the result in metadata.name.
func (c *ComponentSet) MakeAndSetName() {
	c.ObjectMeta.Name = CreateName(c.Spec.SetName, c.Spec.Version)
	return
}

// MakeAndSetName constructs the name from the ComponentPackage's ComponentName
// and Version and stores the result in metadata.name.
func (c *ComponentPackage) MakeAndSetName() {
	c.ObjectMeta.Name = CreateName(c.Spec.ComponentName, c.Spec.Version)
	return
}

// ComponentReference creates a ComponentReference from a component.
func (c *ComponentPackage) ComponentReference() ComponentReference {
	return ComponentReference{
		ComponentName: c.Spec.ComponentName,
		Version:       c.Spec.Version,
	}
}

// ComponentSet creates a ComponentSet from a Bundle. Only components that
// are inlined into the Bundle are considered for the purposes of ComponentSet
// creation.
func (b *Bundle) ComponentSet() *ComponentSet {
	cset := &ComponentSet{
		Spec: ComponentSetSpec{
			SetName: b.SetName,
			Version: b.Version,
		},
	}
	for _, comp := range b.Components {
		cset.Spec.Components = append(cset.Spec.Components, comp.ComponentReference())
	}
	return cset
}

// MakeAndSetName constructs the name from the Bundle's SetName and Version
// and stores the result in metadata.name.
func (b *Bundle) MakeAndSetName() {
	b.ObjectMeta.Name = CreateName(b.SetName, b.Version)
	return
}

// MakeAndSetAllNames constructs the metadata.name for the Bundle and all the
// child inlined components.
func (b *Bundle) MakeAndSetAllNames() {
	b.MakeAndSetName()
	for _, comp := range b.Components {
		comp.MakeAndSetName()
	}
	return
}

// ParsedURL parses the URL in a file object. If no scheme is present, the
// scheme is assumed to be a filepath on the local filesystem.
func (f File) ParsedURL() (*url.URL, error) {
	u := f.URL
	if u == "" {
		return nil, fmt.Errorf("file %v was specified but no URL was provided", f)
	}
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %q: %v", u, err)
	}
	return parsedURL, nil
}
