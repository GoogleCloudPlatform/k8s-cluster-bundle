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

package filter

import (
	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	log "github.com/golang/glog"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// Filterer filters the components and objects in bundles to produce new,
// smaller bundles
type Filterer struct {
	b *bpb.ClusterBundle
}

// Options for filtering bundles. By default, if any of the options match, then
// the relevant component or object is removed. If KeepOnly is set, then the
// objects are kept instead of removed.
type Options struct {
	// Kinds represent the Kinds to filter on.
	Kinds []string

	// Names represent the metadata.names to filter on.
	Names []string

	// Annotations contain key/value pairs to filter on. An empty string value matches
	// all annotation-values for a particular key.
	Annotations map[string]string

	// Labels contain key/value pairs to filter on. An empty string value matches
	// all label-values for a particular key.
	Labels map[string]string

	// KeepOnly means that instead of removing the found objects.
	KeepOnly bool
}

// FilterComponents filters components based on the ObjectMeta properties of
// the components, returning a new cluster bundle with just filtered
// components. By default components are removed, unless KeepOnly is set, and
// then the opposite is true. Filtering for components doesn't take into
// account the properties of the object-children of the components.
func FilterComponents(b *bpb.ClusterBundle, o *Options) *bpb.ClusterBundle {
	b = converter.CloneBundle(b)
	if b.GetSpec() == nil {
		return b
	}
	var matched []*bpb.ClusterComponent
	var notMatched []*bpb.ClusterComponent
	for _, c := range b.GetSpec().GetComponents() {
		matches := filterMeta(c.GetKind(), c.GetMetadata(), o)
		if matches {
			matched = append(matched, c)
		} else {
			notMatched = append(notMatched, c)
		}
	}
	if o.KeepOnly {
		b.GetSpec().Components = matched
		return b
	}
	b.GetSpec().Components = notMatched
	return b
}

// FilterObjects filters objects based on the ObjectMeta properties of
// the objects, returning a new cluster bundle with just filtered
// objects. By default objectsare removed, unless KeepOnly is set, and
// then the opposite is true.
func FilterObjects(b *bpb.ClusterBundle, o *Options) *bpb.ClusterBundle {
	b = converter.CloneBundle(b)
	if b.GetSpec() == nil {
		return b
	}
	for _, cp := range b.GetSpec().GetComponents() {
		var matched []*structpb.Struct
		var notMatched []*structpb.Struct
		for _, c := range cp.GetClusterObjects() {
			meta, err := converter.ObjectMetaFromStruct(c)
			if err != nil {
				log.Infof("error converting cluster's object meta: %v", err)
				// If this happens, then likely, the structure is invalid of the
				// ObjectMeta.
				continue
			}
			matches := filterMeta(c.GetFields()["kind"].GetStringValue(), meta, o)
			if matches {
				matched = append(matched, c)
			} else {
				notMatched = append(notMatched, c)
			}
		}
		if o.KeepOnly {
			cp.ClusterObjects = matched
		} else {
			cp.ClusterObjects = notMatched
		}
	}
	return b
}

func filterMeta(kind string, meta *bpb.ObjectMeta, o *Options) bool {
	for _, k := range o.Kinds {
		if k == kind {
			return true
		}
	}
	// We ignore kind, because Kind should always be cluster bundle.
	for _, n := range o.Names {
		if meta.GetName() == n {
			return true
		}
	}
	if len(meta.Annotations) > 0 {
		for key, v := range o.Annotations {
			if val, ok := meta.Annotations[key]; ok && val == v {
				return true
			}
		}
	}
	if len(meta.Labels) > 0 {
		for key, v := range o.Labels {
			if val, ok := meta.Labels[key]; ok && val == v {
				return true
			}
		}
	}
	return false
}
