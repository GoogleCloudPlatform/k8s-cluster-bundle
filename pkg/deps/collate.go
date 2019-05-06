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
	out += "["
	for i, v := range s.versions {
		if i != 0 {
			out += " "
		}
		out += v.version.String()
	}
	return out + "] }"
}

// latest gets the latest depMeta version for a component, given some
// annotation criteria, or returns an error if no latest version is
// available.
func (s *sortedVersions) latest(matcher Matcher) (*depMeta, error) {
	for i := len(s.versions) - 1; i >= 0; i-- {
		ver := s.versions[i]
		if matcher == nil || matcher(ver.ref(), ver.matchMeta) {
			return ver, nil
		}
	}
	return nil, fmt.Errorf("for component %s, no latest version", s.componentName)
}

// Previous gets the version previous to a specific version, given some
// annotation criteria, or returns an error if no previous version is
// available.
func (s *sortedVersions) previous(ver semver.Version, matcher Matcher) (*depMeta, error) {
	// This is currently a linear search for clarity. Given the filtering
	// options, it's not clear if it could be made significantly faster.
	for i := len(s.versions) - 1; i >= 0; i-- {
		m := s.versions[i]
		if m.version.LT(ver) && (matcher == nil || matcher(m.ref(), m.matchMeta)) {
			return m, nil
		}
	}
	return nil, fmt.Errorf("for component %s, no previous version to %v", s.componentName, ver)
}

// sortedMapFromComponents creates a map from component name to meta.
func sortedMapFromComponents(comps []*bundle.Component, mp MatchProcessor) (map[string]*sortedVersions, map[bundle.ComponentReference]*depMeta, error) {
	allMetas := make(map[string][]*depMeta)
	lookupMap := make(map[bundle.ComponentReference]*depMeta)
	for _, c := range comps {
		m, err := metaFromComponent(c, mp)
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
		orderByVersion(metas)
		sortedMap[name] = &sortedVersions{
			componentName: name,
			versions:      metas,
		}
	}
	return sortedMap, lookupMap, nil
}

// orderByVersion sorts the meta objects by version, in descending order
// (latest first).
//
// This method assumes that the list of metas is already been
// collated so it contains just the metas for a single component.
func orderByVersion(m []*depMeta) {
	sort.Slice(m, func(i, j int) bool {
		return m[i].version.LT(m[j].version)
	})
}
