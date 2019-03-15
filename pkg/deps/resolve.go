// Copyright 2019 Google LLC
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

// Package deps resolves component dependencies based on a collection of
// components and some input-requirements
package deps

import (
	"fmt"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// Resolver is a dependency resolver.
type Resolver struct {
	componentVersions map[string]*sortedVersions
	metaLookup        map[bundle.ComponentReference]*depMeta
	componentLookup   map[bundle.ComponentReference]*bundle.Component
}

// NewResolver takes in a list of components that make up the universe.
func NewResolver(components []*bundle.Component) (*Resolver, error) {
	svm, lup, err := sortedMapFromComponents(components)
	if err != nil {
		return nil, err
	}
	componentLookup := make(map[bundle.ComponentReference]*bundle.Component)
	for _, c := range components {
		componentLookup[c.ComponentReference()] = c
	}

	return &Resolver{
		componentVersions: svm,
		metaLookup:        lup,
		componentLookup:   componentLookup,
	}, nil
}

// Component retrieves a component (which is copied for safety), returning nil
// if isn't known by the resolver.
func (r *Resolver) Component(ref bundle.ComponentReference) *bundle.Component {
	c := r.componentLookup[ref]
	if c == nil {
		return nil
	}
	return c.DeepCopy()
}

// HasComponent indicates whether a component is known by the resolver.
func (r *Resolver) HasComponent(ref bundle.ComponentReference) bool {
	return r.componentLookup[ref] != nil
}

// Add adds components to the resolver, returning a new resolver or
// an error if it fails. If one of the new components provided has the same
// name/version, it overwrites the existing component map.
func (r *Resolver) Add(newComps []*bundle.Component) (*Resolver, error) {
	m := make(map[bundle.ComponentReference]*bundle.Component)
	for k, c := range r.componentLookup {
		m[k] = c
	}
	for _, c := range newComps {
		m[c.ComponentReference()] = c
	}
	var cs []*bundle.Component
	for _, c := range m {
		cs = append(cs, c)
	}
	return NewResolver(cs)
}

// ComponentVersions gets the versions for a particular component
func (r *Resolver) ComponentVersions(comp string) ([]bundle.ComponentReference, error) {
	cv, ok := r.componentVersions[comp]
	if !ok {
		return nil, fmt.Errorf("unknown component %q", comp)
	}
	var versions []bundle.ComponentReference
	for _, ver := range cv.versions {
		versions = append(versions, bundle.ComponentReference{
			ComponentName: cv.componentName,
			Version:       ver.version.String(),
		})
	}
	return versions, nil
}

// ResolveOptions are options for resolving dependencies
type ResolveOptions struct {
	// Match are component annotations that must to be present. This is useful
	// for matching positive characteristics of a component (example: this
	// component passed qualification). Only one of the list of values need
	// be present.
	//
	// In other words, this is a logical AND operation of the passed in Annotation and
	// the component Annotation.
	Match map[string][]string

	// MatchIfPresent are component annotations that are analyzed if they are
	// present on the component. This is useful for synchronizing values between
	// the universe of components and the caller.
	MatchIfPresent map[string][]string

	// Exclude are component annotations that must not be present.  This is
	// useful for matching positive characteristics of a component (this
	// component passed qualification). If any of the the list of values match,
	// the component is excluded.
	//
	// In other words, this is a logical NAND operation of the passed in
	// Annotation and the component Annotation.
	Exclude map[string][]string
}

// ResolveLatest resolves dependencies for several components, returning
// the components references that correspond to that version selection.
//
// These components must exist within the Resolver's set of components. If a
// version is not specified for a component reference, the latest version is
// used.
func (r *Resolver) ResolveLatest(refs []bundle.ComponentReference, opts *ResolveOptions) ([]bundle.ComponentReference, error) {
	var exact []*depMeta
	var latest []*depMeta
	sopts := &searchOpts{}
	if opts != nil {
		sopts.match = opts.Match
		sopts.exclude = opts.Exclude
		sopts.matchIfPresent = opts.MatchIfPresent
	}
	for _, ref := range refs {
		cv, ok := r.componentVersions[ref.ComponentName]
		if !ok {
			return nil, fmt.Errorf("unknown component %q", ref.ComponentName)
		}
		if cv == nil || len(cv.versions) == 0 {
			return nil, fmt.Errorf("no versions found for component %q", ref.ComponentName)
		}
		if ref.Version == "" {
			ver, err := cv.latest(sopts)
			if err != nil {
				return nil, err
			}
			latest = append(latest, ver)
		} else {
			m, ok := r.metaLookup[ref]
			if !ok {
				versions, err := r.ComponentVersions(ref.ComponentName)
				if err != nil {
					return nil, err
				}
				return nil, fmt.Errorf("unknown component %v; known versions are %v", ref, versions)
			}
			exact = append(exact, m)
		}
	}
	return r.findLatest(exact, latest, sopts)
}
