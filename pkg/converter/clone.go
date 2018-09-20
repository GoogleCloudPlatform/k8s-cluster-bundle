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

package converter

import (
	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// CloneBundle creates a copy of a cluster bundle proto.
func CloneBundle(b *bpb.ClusterBundle) *bpb.ClusterBundle {
	return proto.Clone(b).(*bpb.ClusterBundle)
}

// CloneClusterComponent creates a copy of a object collection proto.
func CloneClusterComponent(a *bpb.ClusterComponent) *bpb.ClusterComponent {
	return proto.Clone(a).(*bpb.ClusterComponent)
}

// CloneStruct creates a copy of a struct proto.
func CloneStruct(b *structpb.Struct) *structpb.Struct {
	return proto.Clone(b).(*structpb.Struct)
}
