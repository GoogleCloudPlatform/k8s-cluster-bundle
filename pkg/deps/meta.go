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
	"regexp"
	"strings"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/blang/semver"
)

// A single requested dependency.
type requestedDep struct {
	componentName string
	version       semver.Version
}

// String formats the requestedDep as a string
func (r *requestedDep) String() string {
	return fmt.Sprintf("{%s, %s}", r.componentName, r.version)
}

// depMeta contains dependency metadata about a single component.
type depMeta struct {
	componentName string
	version       semver.Version
	reqDeps       []*requestedDep
	visibility    map[string]bool
	annotations   map[string]string
}

// String prints a stringified version of the depMeta object.
func (m *depMeta) String() string {
	baseStr := fmt.Sprintf("{reference:{%s, %s}", m.componentName, m.version.String())
	if len(m.reqDeps) > 0 {
		baseStr = baseStr + fmt.Sprintf(" requirements:%v", m.reqDeps)
	}
	if len(m.visibility) > 0 {
		baseStr = baseStr + fmt.Sprintf(" visibility:%v", m.visibility)
	}
	baseStr += "}"
	return baseStr
}

// versionStr returns the version string.
func (m *depMeta) versionStr() string { return m.version.String() }

// ref makes a component reference.
func (m *depMeta) ref() bundle.ComponentReference {
	return bundle.ComponentReference{m.componentName, m.version.String()}
}

// visibleTo returns whether this component is visible to another component.
func (m *depMeta) visibleTo(other *depMeta) bool {
	if len(m.visibility) == 0 || m.visibility["@private"] {
		return false
	}
	if m.visibility["@public"] {
		return true
	}
	return m.visibility[other.componentName]
}

// metaFromComponent creates the depMeta object from a component.
func metaFromComponent(c *bundle.Component) (*depMeta, error) {
	if c.Spec.ComponentName == "" || c.Spec.Version == "" {
		return nil, fmt.Errorf("both componentName and version must be defined for each component, but found componentName:%s, version:%s", c.Spec.ComponentName, c.Spec.Version)
	}

	m := &depMeta{
		componentName: c.Spec.ComponentName,
		visibility:    make(map[string]bool),
		annotations:   make(map[string]string),
	}

	ver, err := semver.Parse(c.Spec.Version)
	if err != nil {
		return nil, fmt.Errorf("while parsing version %q as semver in component %v: %v", c.Spec.Version, c.ComponentReference(), err)
	}
	m.version = ver

	if len(c.ObjectMeta.Annotations) > 0 {
		for k, v := range c.ObjectMeta.Annotations {
			m.annotations[k] = v
		}
	}

	var req *bundle.Requirements
	for _, o := range c.Spec.Objects {
		if o.GetKind() != "Requirements" {
			continue
		}
		if o.GetAPIVersion() != "" && !strings.Contains(o.GetAPIVersion(), "bundle.gke.io") {
			continue
		}
		inreq := &bundle.Requirements{}
		err := converter.FromUnstructured(o).ToObject(inreq)
		if err != nil {
			return nil, fmt.Errorf("for component %v: %v", m.ref(), err)
		}
		if req != nil {
			return nil, fmt.Errorf("duplicate requirements object found for component %v", m.ref())
		}
		req = inreq
	}

	if req != nil {
		for _, vis := range req.Visibility {
			if vis != "" {
				m.visibility[vis] = true
			}
		}
	}

	reqDeps, err := flattenRequired(m.ref(), req)
	if err != nil {
		return nil, err
	}
	m.reqDeps = reqDeps
	return m, nil
}

var (
	// numPattern is a regexp string for numbers without leading zeros.
	numPattern = `([1-9]\d*|0)`

	// semvers of the form X.Y
	majorMinorPattern = regexp.MustCompile(fmt.Sprintf(`^%s\.%s$`, numPattern, numPattern))
)

// flattenRequired takes a requirements object from a component and flattens
// into a list of requested deps.
func flattenRequired(ref bundle.ComponentReference, req *bundle.Requirements) ([]*requestedDep, error) {
	if req == nil {
		// This is a valid/normal case -- no dependencies specified.
		return nil, nil
	}

	var deps []*requestedDep
	for _, d := range req.Require {
		if d.ComponentName == "" {
			return nil, fmt.Errorf("for component %v, component name was not defined for require field %v in requirements object %v", ref, d, req)
		}
		verstr := d.Version
		if verstr == "" {
			verstr = "0.0.0"
		}
		if majorMinorPattern.MatchString(verstr) {
			verstr = verstr + ".0"
		}
		ver, err := semver.Parse(verstr)
		if err != nil {
			return nil, fmt.Errorf("for component %v, error while parsing requirement version %q as semver for required component %q: %v", ref, verstr, d.ComponentName, err)
		}

		deps = append(deps, &requestedDep{
			componentName: d.ComponentName,
			version:       ver,
		})
	}
	return deps, nil
}
