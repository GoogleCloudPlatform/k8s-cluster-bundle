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
	"encoding/json"

	bextpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundleext/v1alpha1"
	structpb "github.com/golang/protobuf/ptypes/struct"
	corev1 "k8s.io/api/core/v1"
)

// KubeConverter converts structs to Kubernetes objects.
type KubeConverter struct {
	s *structpb.Struct
}

// FromStruct creates a new KubeConverter.
func FromStruct(s *structpb.Struct) *KubeConverter {
	return &KubeConverter{s}
}

// ToNodeConfig converts from a struct to a NodeConfig.
func (k *KubeConverter) ToNodeConfig() (*bextpb.NodeConfig, error) {
	b, err := Struct.ProtoToJSON(k.s)
	if err != nil {
		return nil, err
	}
	pb, err := NodeConfig.JSONToProto(b)
	if err != nil {
		return nil, err
	}
	return ToNodeConfig(pb), nil
}

// ToNodeConfig converts from a struct to a NodeConfig.
func (k *KubeConverter) ToConfigMap() (*corev1.ConfigMap, error) {
	b, err := Struct.ProtoToJSON(k.s)
	if err != nil {
		return nil, err
	}

	cm := &corev1.ConfigMap{}
	err = json.Unmarshal(b, cm)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

/*
For now, remove the apiextensions-apiserver dependency. It's not really used anywhere
anyway at the moment

// ToCRD converts a struct to a Kubernetes CustomResourceDefinition.
func (k *KubeConverter) ToCRD() (*apiextv1beta1.CustomResourceDefinition, error) {
	crd := &apiextv1beta1.CustomResourceDefinition{}
	b, err := Struct.ProtoToJSON(k.s)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, crd)
	if err != nil {
		return nil, err
	}

	return crd, nil
}
*/
