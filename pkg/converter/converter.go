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
)

var (
	// Bundle is a converter for ClusterBundles
	Bundle = &Converter{&bpb.ClusterBundle{}}

	// ComponentPackage is a converter for ComponentPackage protos.
	ComponentPackage = &Converter{&bpb.ComponentPackage{}}

	// NodeConfig is a converter for NodeConfig protos.
	NodeConfig = &Converter{&bpb.NodeConfig{}}

	// ObjectMeta is a converter for NodeConfig protos.
	ObjectMeta = &Converter{&bpb.ObjectMeta{}}

	// Struct is a converter for Struct protos.
	Struct = &Converter{&structpb.Struct{}}
)

// ToBundle is a type converter for converting to and from a Bundle.
func ToBundle(msg proto.Message) *bpb.ClusterBundle {
	return msg.(*bpb.ClusterBundle)
}

// ToComponentPackage is a type converter for converting to and from a ComponentPackage.
func ToComponentPackage(msg proto.Message) *bpb.ComponentPackage {
	return msg.(*bpb.ComponentPackage)
}

// ToStruct is a type converter for converting to and from a Struct.
func ToStruct(msg proto.Message) *structpb.Struct {
	return msg.(*structpb.Struct)
}

// ToNodeConfig is a type converter for converting to and from a NodeConfig.
func ToNodeConfig(msg proto.Message) *bpb.NodeConfig {
	return msg.(*bpb.NodeConfig)
}

// ToObjectMeta is a type converter for converting a proto to an ObjectMeta.
func ToObjectMeta(msg proto.Message) *bpb.ObjectMeta {
	return msg.(*bpb.ObjectMeta)
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

// ProtoToYAML converts a proto into a YAML encoded proto.
func (b *Converter) ProtoToYAML(bun proto.Message) ([]byte, error) {
	return ProtoToYAML(bun)
}

// ProtoToYAML converts a proto into a YAML encoded proto.
func ProtoToYAML(b proto.Message) ([]byte, error) {
	buf, err := ProtoToJSON(b)
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
	return ProtoToJSON(bun)
}

// ProtoToJSON converts a proto into a JSON encoded proto.
func ProtoToJSON(b proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := (&jsonpb.Marshaler{}).Marshal(&buf, b); err != nil {
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

// ObjectMetaFromStruct converts to the ObjectMeta proto from a structpb's ObjectMeta.
func ObjectMetaFromStruct(obj *structpb.Struct) (*bpb.ObjectMeta, error) {
	m := &bpb.ObjectMeta{}
	metaf := obj.GetFields()["metadata"]
	if metaf == nil {
		return m, nil
	}

	by, err := Struct.ProtoToJSON(&structpb.Struct{Fields: metaf.GetStructValue().GetFields()})
	if err != nil {
		return nil, fmt.Errorf("error marshalling to json: %v", err)
	}

	objm, err := ObjectMeta.JSONToProto(by)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling to object meta: %v", err)
	}

	return ToObjectMeta(objm), nil
}
