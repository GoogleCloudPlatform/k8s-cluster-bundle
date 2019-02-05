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

package build

import (
	"bytes"
	"fmt"
	"text/template"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/wrapper"
	log "github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComponentFiles reads file-references for component builder objects.
// The returned components are copies with the file-references removed.
func BuildPatchTemplate(ptb *bundle.PatchTemplateBuilder, opts options.JSONOptions) (*bundle.PatchTemplate, error) {
	name := ptb.GetName()
	if ptb.Template == "" {
		return nil, fmt.Errorf("cannot build PatchTemplate from PatchTemplateBuilder %q: it has an empty template", name)
	}

	tmpl, err := template.New("ptb").Parse(ptb.Template)
	if err != nil {
		return nil, fmt.Errorf("cannot build PatchTemplate from PatchTemplateBuilder %q: error parsing template: %v", name, err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, opts)
	if err != nil {
		return nil, fmt.Errorf("cannot build PatchTemplate from PatchTemplateBuilder %q: error executing template: %v", name, err)
	}

	pt := &bundle.PatchTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "bundle.gke.io/v1alpha1",
			Kind:       "PatchTemplate",
		},
		ObjectMeta:    *ptb.ObjectMeta.DeepCopy(),
		OptionsSchema: ptb.TargetSchema.DeepCopy(),
		Template:      buf.String(),
	}
	return pt, nil
}

func BuildComponentPatchTemplates(c *bundle.Component, fopts *filter.Options, opts options.JSONOptions) (*bundle.Component, error) {
	log.Info("Building PatchTemplates for component")

	ptbFilter := fopts
	if ptbFilter == nil {
		ptbFilter = &filter.Options{}
	}
	ptbFilter.Kinds = []string{`PatchTemplateBuilder`}

	f := filter.NewFilter()
	// select all patch template builders that match the filter
	ptbs, objs := f.PartitionObjects(c.Spec.Objects, ptbFilter)
	c.Spec.Objects = objs

	// loop through each and generate a patch template
	for _, obj := range ptbs {
		var ptb bundle.PatchTemplateBuilder
		err := converter.FromUnstructured(obj).ToObject(&ptb)
		if err != nil {
			return nil, err
		}

		pt, err := BuildPatchTemplate(&ptb, opts)
		if err != nil {
			return nil, err
		}
		y, err := converter.FromObject(pt).ToYAML()
		if err != nil {
			return nil, err
		}
		o, err := converter.FromYAML(y).ToUnstructured()
		if err != nil {
			return nil, err
		}
		c.Spec.Objects = append(c.Spec.Objects, o)
	}

	return c, nil
}

func BuildAllPatchTemplates(bw *wrapper.BundleWrapper, fopts *filter.Options, opts options.JSONOptions) (*wrapper.BundleWrapper, error) {
	switch bw.Kind() {
	case "Component":
		comp, err := BuildComponentPatchTemplates(bw.Component(), fopts, opts)
		if err != nil {
			return nil, err
		}
		bw = wrapper.FromComponent(comp)
	case "Bundle":
		log.Info("Building PatchTemplates for bundle")
		bun := bw.Bundle()
		var comps []*bundle.Component
		for _, comp := range bun.Components {
			comp, err := BuildComponentPatchTemplates(comp, fopts, opts)
			if err != nil {
				return nil, err
			}
			comps = append(comps, comp)
		}
		bun.Components = comps
		bw = wrapper.FromBundle(bun)
	default:
		return nil, fmt.Errorf("bundle kind %q not supported for patching", bw.Kind())
	}

	return bw, nil
}
