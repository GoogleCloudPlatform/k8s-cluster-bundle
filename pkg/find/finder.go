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

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

// BundleFinder is a wrapper which allows for efficient searching through
// bundles. The BundleFinder is intended to be readonly; if modifications are
// made to the bundle, subsequent lookups will fail.
type BundleFinder struct {
	bundle        *bpb.ClusterBundle
	nodeLookup    map[string]*bpb.NodeConfig
	compLookup    map[string]*bpb.ClusterComponent
	compObjLookup map[core.ClusterObjectKey]*bpb.ClusterObject
}

// NewBundleFinder creates a new BundleFinder or returns an error.
func NewBundleFinder(b *bpb.ClusterBundle) (*BundleFinder, error) {
	b = converter.CloneBundle(b)
	// TODO: we assume the bundle is in a correct state at this point.
	// should we? Should we validate here?
	nodeConfigs := make(map[string]*bpb.NodeConfig)
	for _, nc := range b.GetSpec().GetNodeConfigs() {
		n := nc.GetName()
		if n == "" {
			return nil, fmt.Errorf("node bootstrap configs must always have a name. was empty for %v", nc)
		}
		nodeConfigs[n] = nc
	}

	compConfigs := make(map[string]*bpb.ClusterComponent)
	compObjLookup := make(map[core.ClusterObjectKey]*bpb.ClusterObject)
	for _, ca := range b.GetSpec().GetComponents() {
		n := ca.GetName()
		if n == "" {
			return nil, fmt.Errorf("cluster components must always have a name. was empty for %v", ca)
		}
		compConfigs[n] = ca
		for _, co := range ca.GetClusterObjects() {
			con := co.GetName()
			if con == "" {
				return nil, fmt.Errorf("cluster component objects must always have a name. was empty for object %v in component %q", co, n)
			}
			compObjLookup[core.ClusterObjectKey{n, con}] = co
		}
	}

	return &BundleFinder{
		bundle:        b,
		nodeLookup:    nodeConfigs,
		compLookup:    compConfigs,
		compObjLookup: compObjLookup,
	}, nil
}

// ClusterComponent returns a found cluster component or nil.
func (b *BundleFinder) ClusterComponent(name string) *bpb.ClusterComponent {
	return b.compLookup[name]
}

// NodeConfig returns a node bootstrap config or nil.
func (b *BundleFinder) NodeConfig(name string) *bpb.NodeConfig {
	return b.nodeLookup[name]
}

// ClusterComponentObject returns a ClusterComponent's Cluster object or nil.
func (b *BundleFinder) ClusterComponentObject(compName string, objName string) *bpb.ClusterObject {
	return b.compObjLookup[core.ClusterObjectKey{compName, objName}]
}
