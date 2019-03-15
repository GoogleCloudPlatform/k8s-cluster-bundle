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

	// Maximum of the min requirements placed on this node.
	maxOfMinReq semver.Version

	// fixed determines whether this is fixed, meaning it cannot be downgraded.
	// In other words, this is one of the components initially passed-in.
	fixed bool
}

// String formats the data for the current node.
func (n *node) String() string {
	return fmt.Sprintf("{meta:%v, fixed:%t, maxOfMinReq:%v}", n.meta, n.fixed, n.maxOfMinReq)
}

func (r *Resolver) findLatest(exact, latest []*depMeta, opts *searchOpts) ([]bundle.ComponentReference, error) {
	root := &node{}
	toVisit := []*node{}
	inQueue := make(map[string]bool)

	initAdd := func(m *depMeta, isExact bool) {
		n := &node{
			meta:  m,
			fixed: isExact,
		}
		root.children = append(root.children, n)
		toVisit = append(toVisit, n)
		inQueue[n.meta.componentName] = true
	}

	for _, m := range exact {
		initAdd(m, true)
	}
	for _, m := range latest {
		initAdd(m, false)
	}

	graphBuilder := &graphBuilder{
		searchOpts: opts,
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
		for _, c := range children {
			_, alreadyPicked := graphBuilder.picked[c.meta.componentName]
			_, alreadyInQueue := inQueue[c.meta.componentName]
			if !alreadyPicked && !alreadyInQueue {
				toVisit = append(toVisit, c)
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
	searchOpts        *searchOpts
}

func shiftNode(ns []*node) (*node, []*node) {
	return ns[0], ns[1:]
}

// visitNode constructs the latest graph for a single node. It returns the
// nodes that need to be visited next or an error if things didn't work out.
func (r *graphBuilder) visitNode(curNode *node) ([]*node, error) {
	for success := false; !success; {
		// We'll assume that adding the deps went swimmingly.
		success = true

		for _, dep := range curNode.meta.reqDeps {
			picked, alreadyPicked := r.picked[dep.componentName]

			var newChild *node
			if alreadyPicked {
				// We already picked a component with the same component name. Check to
				// make sure the version works with this component's dependency
				// requirements
				successfulPick, err := r.processPicked(dep, curNode, picked)
				if err != nil {
					return nil, err
				}
				if !successfulPick {
					success = false
					break
				}
				newChild = picked
			} else {
				// Do selection for the first time. We'll pick the latest component
				// version of the component that satisfies the requirements.
				latest, err := r.pickLatest(dep, curNode)
				if err != nil {
					return nil, err
				}
				if latest == nil {
					success = false
					break
				}
				newChild = latest
			}
			if !newChild.meta.visibleTo(curNode.meta) {
				return nil, fmt.Errorf("component %v is not visible to component %v", newChild.meta, curNode.meta)
			}
			curNode.children = append(curNode.children, newChild)
		}

		if !success {
			err := r.tryDowngrade(curNode)
			if err != nil {
				return nil, err
			}
		}
	}
	return curNode.children, nil
}

// processPicked checks to make sure an already picked component works with the
// existing constraints
func (r *graphBuilder) processPicked(dep *requestedDep, curNode, picked *node) (bool, error) {
	if picked.meta.version.GTE(dep.version) {
		// It works!
		return true, nil
	}
	if dep.version.GT(picked.maxOfMinReq) {
		picked.maxOfMinReq = dep.version
	}
	// It doesn't work. Indicate to the caller that a downgrade needs to be performed.
	return false, nil
}

// pickLatest picks the latest component for a given dependency requirement.
func (r *graphBuilder) pickLatest(dep *requestedDep, curNode *node) (*node, error) {
	sorted, ok := r.componentVersions[dep.componentName]
	if !ok {
		return nil, fmt.Errorf("could not find sorted component versions for component name %q", dep.componentName)
	}
	latest, err := sorted.latest(r.searchOpts)
	if err != nil {
		return nil, err
	}
	if latest.version.LT(dep.version) {
		// This is a special error case: the latest version version of the dep
		// doesn't work with the curNode
		//
		// If the latest version of the dependency is less than the requested
		// version, our only option is to try to downgrade the parent node. We pass
		// back nil to indicate as such.
		return nil, nil
	}
	newChild := &node{
		meta:        latest,
		maxOfMinReq: dep.version,
	}
	return newChild, nil
}

// tryDowngrade tries to downgrade the current node version, in-place.
//
// This is done in the case where there was some, possibly recoverable
// incomptibility. Because we always select the latest version of dependencies,
// we don't have the option to upgrade the picked version. Instead, we can try
// to downgrade current node and start the dependency process over for the
// current node.
func (r *graphBuilder) tryDowngrade(curNode *node) error {
	if curNode.fixed {
		// The current node is fixed because it was one of the components
		// originally passed in. We're stuck.
		return fmt.Errorf("attempting to downgrade component %v, but component was fixed (explicitly specified as one of the initial components)", curNode.meta.ref())
	}
	sorted, ok := r.componentVersions[curNode.meta.componentName]
	if !ok {
		return fmt.Errorf("could not find component versions for component name %q", curNode.meta.componentName)
	}
	previous, err := sorted.previous(curNode.meta.version, r.searchOpts)
	if err != nil {
		return fmt.Errorf("while trying to downgrade component %v, got error: %v", curNode.meta, err)
	}
	if previous.version.LT(curNode.maxOfMinReq) {
		return fmt.Errorf("while trying to downgrade component %v, found incompatibility; previous version %v is less than the min version %v required by other components",
			curNode.meta, previous, curNode.maxOfMinReq)
	}

	// Clear out the children and try again.
	curNode.children = nil
	curNode.meta = previous
	return nil
}
