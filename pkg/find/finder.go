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
)

// Key representing a specific cluster object.
type appObjKey struct {
	appName string
	objName string
}

// BundleFinder is a wrapper which allows for efficient searching through
// bundles. The BundleFinder is intended to be readonly; if modifications are
// made to the bundle, subsequent lookups will fail.
type BundleFinder struct {
	bundle       *bpb.ClusterBundle
	nodeLookup   map[string]*bpb.ImageConfig
	appLookup    map[string]*bpb.ClusterApplication
	appObjLookup map[appObjKey]*bpb.ClusterObject
}

// NewBundleFinder creates a new BundleFinder or returns an error.
func NewBundleFinder(b *bpb.ClusterBundle) (*BundleFinder, error) {
	b = converter.CloneBundle(b)
	// TODO: we assume the bundle is in a correct state at this point.
	// should we? Should we validate here?
	nodeConfigs := make(map[string]*bpb.ImageConfig)
	for _, nc := range b.GetSpec().GetImageConfigs() {
		n := nc.GetName()
		if n == "" {
			return nil, fmt.Errorf("node bootstrap configs must always have a name. was empty for %v", nc)
		}
		nodeConfigs[n] = nc
	}

	appConfigs := make(map[string]*bpb.ClusterApplication)
	appObjLookup := make(map[appObjKey]*bpb.ClusterObject)
	for _, ca := range b.GetSpec().GetClusterApps() {
		n := ca.GetName()
		if n == "" {
			return nil, fmt.Errorf("cluster applications must always have a name. was empty for %v", ca)
		}
		appConfigs[n] = ca
		for _, co := range ca.GetClusterObjects() {
			con := co.GetName()
			if con == "" {
				return nil, fmt.Errorf("cluster application objects must always have a name. was empty for object %v in app %q", co, n)
			}
			appObjLookup[appObjKey{n, con}] = co
		}
	}

	return &BundleFinder{
		bundle:       b,
		nodeLookup:   nodeConfigs,
		appLookup:    appConfigs,
		appObjLookup: appObjLookup,
	}, nil
}

// ClusterApp returns a found cluster application or nil.
func (b *BundleFinder) ClusterApp(name string) *bpb.ClusterApplication {
	return b.appLookup[name]
}

// ImageConfig returns a node bootstrap config or nil.
func (b *BundleFinder) ImageConfig(name string) *bpb.ImageConfig {
	return b.nodeLookup[name]
}

// ClusterAppObject returns a Cluster Application's Kubernetes object or nil.
func (b *BundleFinder) ClusterAppObject(appName string, objName string) *bpb.ClusterObject {
	return b.appObjLookup[appObjKey{appName, objName}]
}
