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
func NewResolver(components []*bundle.Component, mp MatchProcessor) (*Resolver, error) {
	if mp == nil {
		mp = NoOpMatchProcessor
	}
	svm, lup, err := sortedMapFromComponents(components, mp)
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

// AllComponents returns all the componenst known by the resolver
func (r *Resolver) AllComponents() []*bundle.Component {
	var out []*bundle.Component
	for _, val := range r.componentLookup {
		out = append(out, val.DeepCopy())
	}
	return out
}

// ComponentVersions gets the versions for a particular component, returning an
// empty list of component references if the component cannot be found.
func (r *Resolver) ComponentVersions(comp string) []bundle.ComponentReference {
	cv, ok := r.componentVersions[comp]
	if !ok {
		return nil
	}
	var versions []bundle.ComponentReference
	for _, ver := range cv.versions {
		versions = append(versions, bundle.ComponentReference{
			ComponentName: cv.componentName,
			Version:       ver.version.String(),
		})
	}
	return versions
}

// ResolveOptions are options for resolving dependencies
type ResolveOptions struct {
	// Matcher is a boolean matcher for applying additional unconditional criteria to
	// componenst during the pick-process.
	Matcher Matcher
}

// Resolve resolves dependencies for several components, returning the
// components references that correspond to that version selection. In general,
// the latest versions of components are preferred
//
// These components must exist within the Resolver's set of components. If a
// version is not specified for a component reference, the latest version is
// used.
func (r *Resolver) Resolve(refs []bundle.ComponentReference, opts *ResolveOptions) ([]bundle.ComponentReference, error) {
	var exact []*depMeta
	var latest []*depMeta
	if opts == nil {
		opts = &ResolveOptions{}
	}

	matcher := opts.Matcher
	if matcher == nil {
		matcher = NoOpMatcher
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
			ver, err := cv.latest(matcher)
			if err != nil {
				return nil, err
			}
			latest = append(latest, ver)
		} else {
			m, ok := r.metaLookup[ref]
			if !ok {
				return nil, fmt.Errorf("unknown component %v; known versions are %v", ref, r.ComponentVersions(ref.ComponentName))
			}
			exact = append(exact, m)
		}
	}
	return r.findLatest(exact, latest, matcher)
}
