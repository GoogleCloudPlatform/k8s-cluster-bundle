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
	"errors"
	"fmt"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/patch"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// patcher provides an interface for patching overlays into a ClusterBundle for
// dependency injection.
type patcher interface {
	PatchBundle(customResources []map[string]interface{}) (*bpb.ClusterBundle, error)
	PatchComponent(component *bpb.ClusterComponent, customResources []map[string]interface{}) (*bpb.ClusterComponent, error)
}

// componentFinder is an interface for finding a cluster component in a bundle.
type componentFinder interface {
	ClusterComponent(name string) *bpb.ClusterComponent
}

// options contain options flags for the bundle patch command.
type options struct {
	bundlePath string
	optionsCRs []string
	component  string
	output     string
}

// opts is a global options flags instance for reference via the cobra command
// installation.
var opts = &options{}

func validateBaseFlags() error {
	if opts.bundlePath == "" {
		return errors.New("a bundle yaml file must be specified")
	}
	if opts.optionsCRs == nil {
		return errors.New("at least one yaml file for the options custom resources")
	}
	return nil
}

func bundleAction(ctx context.Context, cmd *cobra.Command, _ []string) {
	if err := validateBaseFlags(); err != nil {
		cmdlib.ExitWithHelp(cmd, err.Error())
	}

	if err := runPatchBundle(ctx, opts, &core.LocalFileSystemReaderWriter{}); err != nil {
		log.Exit(err)
	}
}

func runPatchBundle(ctx context.Context, opts *options, reader core.FileReaderWriter) error {
	brw := &converter.BundleReaderWriter{reader}

	b, err := brw.ReadBundleFile(ctx, opts.bundlePath)
	if err != nil {
		return err
	}

	crs, err := readAllOptions(ctx, opts.optionsCRs, reader)
	if err != nil {
		return err
	}

	patcher, err := createPatcherFn(b)
	if err != nil {
		return err
	}

	patched, err := patcher.PatchBundle(crs)
	if err != nil {
		return err
	}
	return brw.WriteBundleFile(ctx, opts.output, patched, cmdlib.DefaultFilePermissions)
}

func componentAction(ctx context.Context, cmd *cobra.Command, _ []string) {
	if err := validateBaseFlags(); err != nil {
		cmdlib.ExitWithHelp(cmd, err.Error())
	}

	if opts.component == "" {
		cmdlib.ExitWithHelp(cmd, "the name of the component to patch must be specified.")
	}

	if err := runPatchComponent(ctx, opts, &core.LocalFileSystemReaderWriter{}); err != nil {
		log.Exit(err)
	}
}

func runPatchComponent(ctx context.Context, opts *options, rw core.FileReaderWriter) error {
	brw := &converter.BundleReaderWriter{rw}

	b, err := brw.ReadBundleFile(ctx, opts.bundlePath)
	if err != nil {
		return err
	}

	crs, err := readAllOptions(ctx, opts.optionsCRs, rw)
	if err != nil {
		return err
	}

	finder, err := createFinderFn(b)
	if err != nil {
		return err
	}

	comp := finder.ClusterComponent(opts.component)
	if comp == nil {
		return fmt.Errorf("could not find component %q in bundle", opts.component)
	}

	patcher, err := createPatcherFn(b)
	if err != nil {
		return err
	}

	patched, err := patcher.PatchComponent(comp, crs)
	if err != nil {
		return err
	}

	yaml, err := converter.Struct.ProtoToYAML(patched)
	if err != nil {
		return err
	}
	return rw.WriteFile(ctx, opts.output, yaml, cmdlib.DefaultFilePermissions)
}

// createFinderFn creates an componentFinder that operates on the given ClusterBundle.
var createFinderFn = func(b *bpb.ClusterBundle) (componentFinder, error) {
	return find.NewBundleFinder(b)
}

// createPatcherFn creates an patcher that operates on the given ClusterBundle.
var createPatcherFn = func(b *bpb.ClusterBundle) (patcher, error) {
	return patch.NewPatcherFromBundle(b)
}

// readAllOptions reads the options custome resources from the given list of
// yaml files. It returns a list of CRs in a map representation, where the CR
// fields are the map keys.
func readAllOptions(ctx context.Context, optionsCRs []string, rw core.FileReaderWriter) ([]map[string]interface{}, error) {
	crs := make([]map[string]interface{}, 0, len(optionsCRs))
	for _, o := range optionsCRs {
		bytes, err := rw.ReadFile(ctx, o)
		if err != nil {
			return nil, fmt.Errorf("error reading options from %s: %v", o, err)
		}
		cr, err := converter.KubeResourceYAMLToMap(bytes)
		if err != nil {
			return nil, err
		}
		crs = append(crs, cr)
	}
	return crs, nil
}
