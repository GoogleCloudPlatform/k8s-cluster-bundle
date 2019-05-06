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

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/blang/semver"
)

// rootNodeName is the name of the root noode. we use @ in the rootNode key
// since @ is not valid chararcter in a componentName.
const rootNodeName = "@root"

// node is a dependency node, which corresponds to a component at a specific
// version. Only one version of a component must exist in the graph.
type node struct {
	// children are nodes that this node depends on.
	children []*node

	meta *depMeta

	// Maximum of the min requirements placed on this node. The idea here is that
	// multiple components may depend on this node. We use this as a way to track
	// the maximum of all the minimum version requirements placed on this node.
	minSatisfyingVersion semver.Version

	// fixed determines whether this is fixed, meaning it cannot be downgraded.
	// In other words, this is one of the components initially passed-in.
	fixed bool
}

// String formats the data for the current node.
func (n *node) String() string {
	return fmt.Sprintf("{meta:%v, fixed:%t, minSatisfyingVersion:%v}", n.meta, n.fixed, n.minSatisfyingVersion)
}

func (r *Resolver) findLatest(exact, latest []*depMeta, matcher Matcher) ([]bundle.ComponentReference, error) {
	root := &node{}
	toVisit := []*node{}
	inQueue := make(map[string]*node)

	initAdd := func(m *depMeta, isExact bool) {
		n := &node{
			meta:  m,
			fixed: isExact,
		}
		root.children = append(root.children, n)
		toVisit = append(toVisit, n)
		inQueue[n.meta.componentName] = n
	}

	for _, m := range exact {
		initAdd(m, true)
	}
	for _, m := range latest {
		initAdd(m, false)
	}

	graphBuilder := &graphBuilder{
		matcher: matcher,
		picked: map[string]*node{
			rootNodeName: root,
		},
		componentVersions: r.componentVersions,
		metaLookup:        r.metaLookup,
	}

	// Do a reverse postorder traversal of all the nodes.
	for len(toVisit) > 0 {
		var curNode *node
		curNode, toVisit = shiftNode(toVisit)
		delete(inQueue, curNode.meta.componentName)

		children, err := graphBuilder.visitNode(curNode)
		if err != nil {
			return nil, err
		}

		curNode.children = children
		for _, c := range children {
			_, alreadyPicked := graphBuilder.picked[c.meta.componentName]
			n, alreadyInQueue := inQueue[c.meta.componentName]
			if !alreadyPicked && !alreadyInQueue {
				if alreadyInQueue {
					// We haven't visited the node yet. However, we want to make sure
					// that we're tracking the minimum version requirements placed on
					// this component.
					if c.minSatisfyingVersion.GT(n.minSatisfyingVersion) {
						n.minSatisfyingVersion = c.minSatisfyingVersion
					}
				} else {
					// This means that some node depends
					toVisit = append(toVisit, c)
				}
			}
		}
		graphBuilder.picked[curNode.meta.componentName] = curNode
	}

	var refs []bundle.ComponentReference
	for name, p := range graphBuilder.picked {
		if name != rootNodeName {
			refs = append(refs, p.meta.ref())
		}
	}
	return refs, nil
}

// graphBuilder constructs a possibly cyclic dependency graph.
type graphBuilder struct {
	picked            map[string]*node
	componentVersions map[string]*sortedVersions
	metaLookup        map[bundle.ComponentReference]*depMeta
	matcher           Matcher
}

func shiftNode(ns []*node) (*node, []*node) {
	return ns[0], ns[1:]
}

// visitNode constructs the latest graph for a single node. It returns the
// nodes that need to be visited next or an error if things didn't work out.
func (r *graphBuilder) visitNode(curNode *node) ([]*node, error) {
	var candidateChildren []*node

	for success := false; !success; {
		// We'll assume that adding the deps went swimmingly.
		success = true

		if curNode.meta.version.LT(curNode.minSatisfyingVersion) {
			return nil, fmt.Errorf("couldn't find version of component %q; the maximum of all the minimum requirements was %v and found no suitable version after that version", curNode.meta.componentName, curNode.minSatisfyingVersion)
		}

		for _, dep := range curNode.meta.reqDeps {
			picked, alreadyPicked := r.picked[dep.componentName]

			var newChild *node
			if alreadyPicked {
				// We already picked a component with the same component name. Check to
				// make sure the version works with this component's dependency
				// requirements
				successfulPick := processPicked(dep, curNode, picked)
				if !successfulPick {
					success = false
					break
				}
				newChild = picked
			} else {
				sorted, ok := r.componentVersions[dep.componentName]
				if !ok {
					return nil, fmt.Errorf("could not find sorted component versions for component name %q", dep.componentName)
				}
				// Do selection for the first time. We'll pick the latest component
				// version of the component that satisfies the requirements.
				latest, err := pickLatest(dep, curNode, sorted, r.matcher)
				if err != nil {
					return nil, err
				}
				if latest.meta.version.LT(dep.version) {
					// This is a special error case: the latest version version of the dep
					// doesn't work with the curNode
					//
					// If the latest version of the dependency is less than the requested
					// version, our only option is to try to downgrade the parent node.
					success = false
					break
				}
				newChild = latest
			}
			if !newChild.meta.visibleTo(curNode.meta) {
				return nil, fmt.Errorf("component %v is not visible to component %v", newChild.meta, curNode.meta)
			}
			candidateChildren = append(candidateChildren, newChild)
		}

		if !success {
			candidateChildren = nil
			sortedVersions, ok := r.componentVersions[curNode.meta.componentName]
			if !ok {
				return nil, fmt.Errorf("couldn't find component versions for component name %q", curNode.meta.componentName)
			}
			previous, err := tryDowngrade(curNode, sortedVersions, r.matcher)
			if err != nil {
				return nil, err
			}
			if previous.version.LT(curNode.minSatisfyingVersion) {
				return nil, fmt.Errorf("while trying to downgrade component %v, found incompatibility; previous version %v is less than the min version %v required by other components",
					curNode.meta, previous, curNode.minSatisfyingVersion)
			}
			curNode.meta = previous
		}
	}
	return candidateChildren, nil
}

// processPicked checks to make sure an already picked component works with the
// existing constraints
func processPicked(dep *requestedDep, curNode, picked *node) bool {
	if picked.meta.version.GTE(dep.version) {
		// It works!
		if dep.version.GT(picked.minSatisfyingVersion) {
			picked.minSatisfyingVersion = dep.version
		}
		return true
	}
	// It doesn't work. Indicate to the caller that a downgrade needs to be performed.
	return false
}

// pickLatest picks the latest component for a given dependency requirement,
// returning a proposed child
func pickLatest(dep *requestedDep, curNode *node, sorted *sortedVersions, matcher Matcher) (*node, error) {
	latest, err := sorted.latest(matcher)
	if err != nil {
		return nil, err
	}
	return &node{
		meta:                 latest,
		minSatisfyingVersion: dep.version,
	}, nil
}

// tryDowngrade tries to downgrade the current node version, in-place.
//
// This is done in the case where there was some, possibly recoverable
// incomptibility. Because we always select the latest version of dependencies,
// we don't have the option to upgrade the picked version. Instead, we can try
// to downgrade current node and start the dependency process over for the
// current node.
func tryDowngrade(curNode *node, sorted *sortedVersions, matcher Matcher) (*depMeta, error) {
	if curNode.fixed {
		// The current node is fixed because it was one of the components
		// originally passed in. We're stuck.
		return nil, fmt.Errorf("attempting to downgrade component %v, but component was fixed (explicitly specified as one of the initial components)", curNode.meta.ref())
	}
	previous, err := sorted.previous(curNode.meta.version, matcher)
	if err != nil {
		return nil, fmt.Errorf("while trying to downgrade component %v, got error: %v", curNode.meta, err)
	}
	return previous, nil
}
