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

package jsonnet

import (
	"errors"
	"fmt"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/openapi"
	"github.com/google/go-jsonnet"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type disableImports struct{}

func (disableImports) Import(importedFrom, importPath string) (jsonnet.Contents, string, error) {
	return jsonnet.Contents{}, "", errors.New("jsonnet imports are disabled")
}

type applier struct {
	importer jsonnet.Importer
}

// NewApplier creates a new optionts applier instance.
func NewApplier(importer jsonnet.Importer) options.Applier {
	if importer == nil {
		importer = disableImports{}
	}
	return &applier{importer}
}

func (m *applier) ApplyOptions(comp *bundle.Component, opts options.JSONOptions) (*bundle.Component, error) {
	comp = comp.DeepCopy()

	matched, notMatched := options.PartitionObjectTemplates(comp.Spec.Objects, string(bundle.TemplateTypeJsonnet))

	newObjs, err := options.ApplyCommon(comp.ComponentReference(), matched, opts, m.applyOptions)
	if err != nil {
		return comp, err
	}
	comp.Spec.Objects = append(notMatched, newObjs...)
	return comp, nil
}

func (a *applier) applyOptions(obj *unstructured.Unstructured, ref bundle.ComponentReference, opts options.JSONOptions) ([]*unstructured.Unstructured, error) {
	// TODO(kashomon): this should probably clone the options for safety.
	objTmpl := &bundle.ObjectTemplate{}
	err := converter.FromUnstructured(obj).ToObject(objTmpl)
	if err != nil {
		return nil, err
	}

	if objTmpl.OptionsSchema != nil {
		opts, err = openapi.ApplyDefaults(opts, objTmpl.OptionsSchema)
		if err != nil {
			return nil, fmt.Errorf("applying schema defaults for object template named %q: %v", obj.GetName(), err)
		}
	}

	optsStr, err := converter.FromObject(opts).ToJSONString()
	if err != nil {
		return nil, fmt.Errorf("unable to searialize JSON options: %v", err)
	}

	vm := jsonnet.MakeVM()
	vm.Importer(a.importer)
	vm.TLACode("opts", optsStr)

	// TODO(mikedanese): derive the snippet path from the ObjectTemplate path if
	// it didn't come from the inliner to make imports happy.
	snippetPath := fmt.Sprintf("%s.jsonnet", ref.ComponentName)
	if objTmpl.ObjectMeta.Annotations != nil && objTmpl.ObjectMeta.Annotations[string(bundle.InlinePathIdentifier)] != "" {
		snippetPath = objTmpl.ObjectMeta.Annotations[string(bundle.InlinePathIdentifier)]
	}
	b, err := vm.EvaluateSnippet(snippetPath, objTmpl.Template)
	if err != nil {
		return nil, err
	}
	var out []*unstructured.Unstructured
	if len(b) > 1 && b[0] == '[' {
		if err := converter.FromJSONString(b).ToObject(&out); err != nil {
			return nil, err
		}
	} else {
		obj, err := converter.FromJSONString(b).ToUnstructured()
		if err != nil {
			return nil, err
		}
		out = append(out, obj)
	}
	return out, err
}
