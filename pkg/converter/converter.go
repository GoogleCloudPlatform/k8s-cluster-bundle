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
	"bytes"
	"fmt"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Bundle is a converter for ClusterBundles
	Bundle = &Converter{&bpb.ClusterBundle{}}

	// ClusterApplication is a converter for ClusterApplication protos.
	ClusterApplication = &Converter{&bpb.ClusterApplication{}}

	// Overlay is a converter for Overlay protos.
	Overlay = &Converter{&bpb.Overlay{}}

	// OverlayCollection is a converter for Overlay protos.
	OverlayCollection = &Converter{&bpb.OverlayCollection{}}

	// Struct is a converter for Struct protos.
	Struct = &Converter{&structpb.Struct{}}
)

// ToBundle is a type converter for converting a proto to a Bundle.
func ToBundle(msg proto.Message) *bpb.ClusterBundle {
	return msg.(*bpb.ClusterBundle)
}

// ToClusterApplication is a type converter for converting a proto to a ClusterApplication.
func ToClusterApplication(msg proto.Message) *bpb.ClusterApplication {
	return msg.(*bpb.ClusterApplication)
}

// ToOverlay is a type converter for converting a proto to an Overlay.
func ToOverlay(msg proto.Message) *bpb.Overlay {
	return msg.(*bpb.Overlay)
}

// ToOverlayCollection is a type converter for converting a proto to an
// OverlayCollection.
func ToOverlayCollection(msg proto.Message) *bpb.OverlayCollection {
	return msg.(*bpb.OverlayCollection)
}

// ToStruct is a type converter for converting a proto to an Struct.
func ToStruct(msg proto.Message) *structpb.Struct {
	return msg.(*structpb.Struct)
}

// ToGVK converts a GroupVersionKind proto to the Kubernetes GroupVersionKind type.
func ToGVK(pb *bpb.GroupVersionKind) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   pb.GetGroup(),
		Version: pb.GetVersion(),
		Kind:    pb.GetKind(),
	}
}

// CustomResourceYAMLToMap converts a Custom Resource YAML to a map of string to interface.
// Custom Resources can have arbitrary fields, and we will not have defined structs for each
// options CR to decouple the existence of options CRs from the bundle library. The YAML will be
// parsed into a map so it allows for accessing arbitrary fields.
// TODO: parse CustomResource into a RawExtension instead of a map.
func CustomResourceYAMLToMap(contents []byte) (map[string]interface{}, error) {
	var cr map[string]interface{}
	err := yaml.Unmarshal([]byte(contents), &cr)
	return cr, err
}

// Converter is a generic struct that knows how to convert between textpb,
// proto, and yamls, for a specific proto message.
type Converter struct {
	Msg proto.Message
}

// TextProtoToProto converts a textformat proto to a proto.
func (b *Converter) TextProtoToProto(textpb []byte) (proto.Message, error) {
	bun := proto.Clone(b.Msg)
	if err := proto.UnmarshalText(string(textpb), bun); err != nil {
		return nil, fmt.Errorf("error unmarshaling bundle: %v", err)
	}
	return bun, nil
}

// ProtoToTextProto converts a proto to a textformat proto.
func (b *Converter) ProtoToTextProto(msg proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := proto.MarshalText(&buf, msg); err != nil {
		return nil, fmt.Errorf("error unmarshaling bundle: %v", err)
	}
	return buf.Bytes(), nil
}

// ProtoToYAML converts a bundle into a YAML encoded bundle
func (b *Converter) ProtoToYAML(bun proto.Message) ([]byte, error) {
	buf, err := b.ProtoToJSON(bun)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(buf)
}

// YAMLToProto converts a yaml encoded bundle into a Proto
func (b *Converter) YAMLToProto(contents []byte) (proto.Message, error) {
	js, err := yaml.YAMLToJSON(contents)
	if err != nil {
		return nil, err
	}
	return b.JSONToProto(js)
}

// ProtoToJSON converts a bundle into a JSON encoded proto.
func (b *Converter) ProtoToJSON(bun proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := (&jsonpb.Marshaler{}).Marshal(&buf, bun); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// JSONToProto converts a json encoded bundle into a proto.
func (b *Converter) JSONToProto(contents []byte) (proto.Message, error) {
	bun := proto.Clone(b.Msg)
	umj := &jsonpb.Unmarshaler{AllowUnknownFields: false}
	if err := umj.Unmarshal(bytes.NewBuffer(contents), bun); err != nil {
		return nil, err
	}
	return bun, nil
}
