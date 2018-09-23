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
	"fmt"
	"regexp"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// KubeConverter converts structs to Kubernetes objects.
type KubeConverter struct {
	s *structpb.Struct
}

// FromStruct creates a new KubeConverter.
func FromStruct(s *structpb.Struct) *KubeConverter {
	return &KubeConverter{s}
}

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

// ToObjectRef converts a Kubernetes object to an object reference.
// TODO(kashomon): deduplicate with ObjectRefFromRawKubeResource
func (k *KubeConverter) ToObjectRef() (core.ObjectReference, error) {
	nullResp := core.ObjectReference{}
	ref := core.ObjectReference{}
	apiVersion := k.s.GetFields()["apiVersion"].GetStringValue()
	if apiVersion == "" {
		return nullResp, fmt.Errorf("no apiVersion field was found for Kubernetes resource %v", k.s)
	}
	matches := regexp.MustCompile("(.+)/(.+)").FindStringSubmatch(apiVersion)
	if matches == nil || len(matches) != 3 {
		return nullResp, fmt.Errorf("Kubernetes resource apiVersion is not formatted as group/version: got %q", apiVersion)
	}
	ref.APIVersion = apiVersion

	kind := k.s.GetFields()["kind"].GetStringValue()
	if kind == "" {
		return nullResp, fmt.Errorf("no kind field was found for Kubernetes resource %v", k.s)
	}
	ref.Kind = kind

	md := k.s.GetFields()["metadata"]
	if md == nil {
		return nullResp, fmt.Errorf("no metadata field was found for Kubernetes resource %v", k.s)
	}

	metadata := md.GetStructValue()
	name := metadata.GetFields()["name"].GetStringValue()
	if name == "" {
		return nullResp, fmt.Errorf("no metadata.name field was found for Kubernetes resource %v", k.s)
	}

	ref.Name = name
	return ref, nil
}
