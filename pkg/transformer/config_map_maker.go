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

package transformer

import (
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

type configMapMaker struct {
	cfgMap *structpb.Struct
}

// Make a new ConfigMap with a metdata.name.
//
// Note that metadata.name fields have restrictions and so passed-in names will
// be sanitized.
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
func newConfigMapMaker(name string) *configMapMaker {
	sanitizedName := converter.SanitizeName(name)
	c := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}

	c.Fields["apiVersion"] = &structpb.Value{Kind: &structpb.Value_StringValue{"v1"}}
	c.Fields["kind"] = &structpb.Value{Kind: &structpb.Value_StringValue{"ConfigMap"}}
	c.Fields["metadata"] = &structpb.Value{Kind: &structpb.Value_StructValue{&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": &structpb.Value{
				Kind: &structpb.Value_StringValue{sanitizedName},
			},
		},
	}}}
	c.Fields["data"] = &structpb.Value{Kind: &structpb.Value_StructValue{&structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}}}

	return &configMapMaker{c}
}

// AddData adds a data-key to the config map.
func (c *configMapMaker) addData(key, value string) {
	data := c.cfgMap.GetFields()["data"]
	dv := data.GetStructValue()

	m := dv.GetFields()

	m[key] = &structpb.Value{Kind: &structpb.Value_StringValue{value}}
}
