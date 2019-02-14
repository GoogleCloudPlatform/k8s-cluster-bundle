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

package patch

import (
	"context"
	"fmt"

	log "github.com/golang/glog"
	"github.com/spf13/cobra"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/patchtmpl"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/wrapper"
)

// options represents options flags for the filter command.
type options struct {
	// patchAnnotations selects a subset of patch templates to apply via annotations.
	// Has the form "foo=bar,biff=baz".
	patchAnnotations string

	// optionsFiles contains yaml or json structured data containing options to
	// apply to PatchTemplates
	optionsFiles []string

	// If keepTemplates is true, PatchTemplates will not be stripped from
	// the component objects.
	keepTemplates bool
}

// opts is a global options instance for reference via the add commands.
var opts = &options{}

func action(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, cmd *cobra.Command, _ []string) {
	gopt := cmdlib.GlobalOptionsValues.Copy()
	brw := cmdlib.NewBundleReaderWriter(fio, sio)
	if err := run(ctx, opts, brw, fio, gopt); err != nil {
		log.Exit(err)
	}
}

func run(ctx context.Context, o *options, brw cmdlib.BundleReaderWriter, rw files.FileReaderWriter, gopt *cmdlib.GlobalOptions) error {
	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading contents: %v", err)
	}

	optData, err := cmdlib.MergeOptions(ctx, rw, o.optionsFiles)
	if err != nil {
		return err
	}

	fopts := &filter.Options{Annotations: cmdlib.ParseStringMap(o.patchAnnotations)}
	applier := patchtmpl.NewApplier(patchtmpl.DefaultPatcherScheme(), fopts, o.keepTemplates)

	switch bw.Kind() {
	case "Component":
		log.Info("Patching component")
		comp, err := applier.ApplyOptions(bw.Component(), optData)
		if err != nil {
			return err
		}
		bw = wrapper.FromComponent(comp)
	case "Bundle":
		log.Info("Patching bundle")
		bun := bw.Bundle()
		var comps []*bundle.Component
		for _, comp := range bun.Components {
			comp, err := applier.ApplyOptions(comp, optData)
			if err != nil {
				return err
			}
			comps = append(comps, comp)
		}
		bun.Components = comps
		bw = wrapper.FromBundle(bun)
	default:
		return fmt.Errorf("bundle kind %q not supported for patching", bw.Kind())
	}

	return brw.WriteBundleData(ctx, bw, gopt)
}
