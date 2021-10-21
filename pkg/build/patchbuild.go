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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/openapi"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/wrapper"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PatchTemplate renders a PatchTemplate from a PatchTemplateBuilder and options.
func PatchTemplate(ptb *bundle.PatchTemplateBuilder, opts options.JSONOptions) (*bundle.PatchTemplate, error) {
	name := ptb.GetName()
	if ptb.Template == "" {
		return nil, fmt.Errorf("cannot build PatchTemplate from PatchTemplateBuilder %q: it has an empty template", name)
	}

	tmpl, err := template.New("ptb").Parse(ptb.Template)
	if err != nil {
		return nil, fmt.Errorf("cannot build PatchTemplate from PatchTemplateBuilder %q: error parsing template: %v", name, err)
	}

	if opts == nil {
		opts = make(options.JSONOptions)
	}

	if ptb.BuildSchema != nil {
		opts, err = openapi.ApplyDefaults(opts, ptb.BuildSchema)
		if err != nil {
			return nil, err
		}
	}

	// This is a hack to allow us to pass through runtime templates variables
	// It is only one level-deep. TODO(jbelamaric): fix it to support nested schema
	if ptb.TargetSchema != nil && ptb.TargetSchema.Properties != nil {
		addParamDefaults("", ptb.TargetSchema.Properties, opts)
	}

	tmpl = tmpl.Option("missingkey=error")
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
		PatchType:     ptb.PatchType,
		ObjectMeta:    *ptb.ObjectMeta.DeepCopy(),
		OptionsSchema: ptb.TargetSchema.DeepCopy(),
		Selector:      ptb.Selector.DeepCopy(),
		Template:      buf.String(),
	}
	return pt, nil
}

// ComponentPatchTemplates iterates through all PatchTemplateBuilders in a
// Components Objects, and converts them into PatchTemplates.
func ComponentPatchTemplates(c *bundle.Component, fopts *filter.Options, opts options.JSONOptions) (*bundle.Component, error) {
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

		pt, err := PatchTemplate(&ptb, opts)
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

// AllPatchTemplates is a convenience method to build all PatchTemplateBuilders into
// PatchTemplates for all Components in a Bundle.
func AllPatchTemplates(bw *wrapper.BundleWrapper, fopts *filter.Options, opts options.JSONOptions) (*wrapper.BundleWrapper, error) {
	switch bw.Kind() {
	case "Component":
		comp, err := ComponentPatchTemplates(bw.Component(), fopts, opts)
		if err != nil {
			return nil, err
		}
		bw = wrapper.FromComponent(comp)
	case "Bundle":
		bun := bw.Bundle()
		var comps []*bundle.Component
		for _, comp := range bun.Components {
			comp, err := ComponentPatchTemplates(comp, fopts, opts)
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

// addParamDefaults is a hack to allow us to pass through runtime templates
// variables, which works by looking at the properties in the target schema in
// a PatchTemplateBuilder. It only works for templates that only use simple
// template variables.
func addParamDefaults(prefix string, props map[string]apiextensions.JSONSchemaProps, op options.JSONOptions) error {
	for k, val := range props {
		tmplKey := prefix + "." + k

		if val.Properties != nil {
			// The schema indicates that the this key-value pair has an object
			// substructure, and so we have to recurse into this structure.
			var nestedOpts options.JSONOptions
			optVal, hasVal := op[k]
			if hasVal && optVal != nil {
				nopt, ok := optVal.(map[string]interface{})
				if !ok {
					return fmt.Errorf("Option value with key %q was expected to be an object-type, but was %v", tmplKey, optVal)
				}
				nestedOpts = nopt
			} else {
				nestedOpts = make(options.JSONOptions)
				op[k] = nestedOpts
			}
			// This property contains nested properties. Recurse into these
			// properties and modify the prefix
			addParamDefaults(tmplKey, val.Properties, nestedOpts)
		} else {
			op[k] = `{{` + tmplKey + `}}`
		}
	}
	return nil
}
