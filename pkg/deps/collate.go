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

package deps

import (
	"fmt"
	"sort"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/blang/semver"
)

// searchOpts are options for searching for latest or previous versions. For a
// more detailed discussion, see the ResolveOptions.
type searchOpts struct {
	match          map[string][]string
	matchIfPresent map[string][]string
	exclude        map[string][]string
}

// String returns the stringified search options.
func (s *searchOpts) String() string {
	return fmt.Sprintf("{%+v, %+v}", s.match, s.exclude)
}

// sortedVersions contain all the component versions for a single component,
// sorted by version. It should be considered immutable and must not be changed
// once created.
type sortedVersions struct {
	componentName string

	// versions contains the versions sorted in ascending order. In otherwords,
	// versions[0] is the lowest version.
	versions []*depMeta
}

// String prints the stringified components.
func (s *sortedVersions) String() string {
	out := fmt.Sprintf("{%s", s.componentName)
	for i, v := range s.versions {
		out += " "
		if i == 0 {
			out += "["
		}
		out += v.version.String()
		if i == len(s.versions)-1 {
			out += "]"
		}
	}
	return out + "}"
}

// useComponent evaluates whether, based on the searchOpts, the component
// should be used.
func (s *sortedVersions) useComponent(m *depMeta, opts *searchOpts) bool {
	if opts == nil {
		return true
	}

	match := true
	for k, vals := range opts.match {
		matchesOne := false
		for i := range vals {
			if m.annotations[k] == vals[i] {
				matchesOne = true
				break
			}
		}
		if !matchesOne {
			match = false
			break
		}
	}

	matchIfPresent := true
	for k, vals := range opts.matchIfPresent {
		matchesOne := false
		for i := range vals {
			if val, ok := m.annotations[k]; !ok || val == vals[i] {
				matchesOne = true
				break
			}
		}
		if !matchesOne {
			matchIfPresent = false
			break
		}
	}

	exclude := false
	for k, vals := range opts.exclude {
		matchesOne := false
		for i := range vals {
			if m.annotations[k] == vals[i] {
				matchesOne = true
				break
			}
		}
		if matchesOne {
			exclude = true
			break
		}
	}
	return match && matchIfPresent && !exclude
}

// latest gets the latest depMeta version for a component, given some
// annotation criteria, or returns an error if no latest version is
// available.
func (s *sortedVersions) latest(opts *searchOpts) (*depMeta, error) {
	for i := len(s.versions) - 1; i >= 0; i-- {
		ver := s.versions[i]
		if s.useComponent(ver, opts) {
			return ver, nil
		}
	}
	return nil, fmt.Errorf("for component %s, no latest version matching criteria %v", s.componentName, opts)
}

// Previous gets the version previous to a specific version, given some
// annotation criteria, or returns an error if no previous version is
// available.
func (s *sortedVersions) previous(ver semver.Version, opts *searchOpts) (*depMeta, error) {
	// This is currently a linear search for clarity. Given the filtering
	// options, it's not clear if it could be made significantly faster.
	for i := len(s.versions) - 1; i >= 0; i-- {
		m := s.versions[i]
		if m.version.LT(ver) && s.useComponent(m, opts) {
			return m, nil
		}
	}
	return nil, fmt.Errorf("for component %s, no previous version to %v matching criteria %v", s.componentName, ver, opts)
}

// sortedMapFromComponents creates a map from component name to meta.
func sortedMapFromComponents(comps []*bundle.Component) (map[string]*sortedVersions, map[bundle.ComponentReference]*depMeta, error) {
	allMetas := make(map[string][]*depMeta)
	lookupMap := make(map[bundle.ComponentReference]*depMeta)
	for _, c := range comps {
		m, err := metaFromComponent(c)
		if err != nil {
			return nil, nil, err
		}
		if _, ok := allMetas[m.componentName]; !ok {
			allMetas[m.componentName] = []*depMeta{m}
		} else {
			allMetas[m.componentName] = append(allMetas[m.componentName], m)
		}
		if lookupMap[c.ComponentReference()] != nil {
			return nil, nil, fmt.Errorf("duplicate component %v", c.ComponentReference())
		}
		lookupMap[c.ComponentReference()] = m
	}

	sortedMap := make(map[string]*sortedVersions)
	for name, metas := range allMetas {
		sortedMap[name] = &sortedVersions{
			componentName: name,
			versions:      orderByVersion(metas),
		}
	}
	return sortedMap, lookupMap, nil
}

// orderByVersion sorts the meta objects by version, in descending order
// (latest first).
//
// This method assumes that the list of metas is already been
// collated so it contains just the metas for a single component.
func orderByVersion(m []*depMeta) []*depMeta {
	m = m[:]
	sort.Slice(m, func(i, j int) bool {
		return m[i].version.LT(m[j].version)
	})
	return m
}
