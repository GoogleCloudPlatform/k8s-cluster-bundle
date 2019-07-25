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

// Package multi is an applier that performs multiple different applier
// approaches.
package multi

import (
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/gotmpl"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/jsonnet"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/patchtmpl"
)

type applier struct {
	appliers []options.Applier
}

// NewApplier creates a new multi applier instance.
func NewApplier(appliers []options.Applier) options.Applier {
	return &applier{appliers: appliers[:]}
}

// NewDefaultApplier creates a default multi-applier
func NewDefaultApplier() options.Applier {
	return NewApplier([]options.Applier{
		gotmpl.NewApplier(),
		jsonnet.NewApplier(nil /* importer */),
		patchtmpl.NewDefaultApplier(),
	})
}

// ApplyOptions by going through the appliers in order.
func (m *applier) ApplyOptions(comp *bundle.Component, opts options.JSONOptions) (*bundle.Component, error) {
	comp = comp.DeepCopy()
	for _, a := range m.appliers {
		c, err := a.ApplyOptions(comp, opts)
		if err != nil {
			return nil, err
		}
		comp = c
	}
	return comp, nil
}
