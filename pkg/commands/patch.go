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

package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/patch"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// patchedFilePermissions is the permission used for the bundle output file (rw-r--r--).
const patchedFilePermissions = os.FileMode(0644)

// Patcher provides an interface for patching overlays into a ClusterBundle.
type Patcher interface {
	PatchBundle(customResources []map[string]interface{}) (*bpb.ClusterBundle, error)
	PatchComponent(component *bpb.ClusterComponent, customResources []map[string]interface{}) (*bpb.ClusterComponent, error)
}

// OptionsReader is an interface for reading options custom resources.
type OptionsReader interface {
	ReadOptions(filepath string) ([]byte, error)
}

// localFileSystemReader implements the OptionsReader interface and reads options custom resources
// from the local filesystem.
type localFileSystemReader struct{}

func (r *localFileSystemReader) ReadOptions(filepath string) ([]byte, error) {
	return ioutil.ReadFile(filepath)
}

// comFinder is an interface for finding a cluster component in a bundle.
type compFinder interface {
	ClusterComponent(name string) *bpb.ClusterComponent
}

type patchOptions struct {
	bundlePath string
	optionsCRs []string
	component  string
	output     string
}

var patchOpts = &patchOptions{}

func addPatchCommand() {
	// patchCmd is the parent patch command, and is unrunnable by itself.
	// Patch subcommands should be added to it.
	patchCmd := &cobra.Command{
		Use:   "patch",
		Short: "Apply patches to a bundle or part of a bundle",
		Long:  "Apply patches for bundle customization provided the given options custom resources. See subcommands for patch usage.",
	}

	// Required patch flags
	// Note: the paths to the bundle and options must be absolute when running with bazel.
	patchCmd.PersistentFlags().StringVarP(&patchOpts.bundlePath, "bundle", "b", "", "The path to the bundle to patch")
	patchCmd.PersistentFlags().StringSliceVarP(&patchOpts.optionsCRs, "options", "p", nil,
		"The yaml files containing the options custom resource(s) (comma separated)")

	// Optional patch flags
	// Note: the output directory path is required when running with bazel, and it must be absolute.
	patchCmd.PersistentFlags().StringVarP(&patchOpts.output, "output", "o", "bundle_patched.yaml",
		"Where to output the patched bundle. By default it outputs to the current working directory")

	patchBundleCmd := &cobra.Command{
		Use:   "bundle",
		Short: "Apply patches to the entire bundle",
		Long:  "Apply all the patches found in a bundle to customize it with the given options custom resources",
		Run:   patchBundleAction,
	}

	patchCompCmd := &cobra.Command{
		Use:   "component",
		Short: "Apply patches to a component in the bundle",
		Long:  "Apply all the patches found in the given component in the bundle to customize it with the given options custom resources",
		Run:   patchCompAction,
	}

	// Required patch component flags
	patchCompCmd.Flags().StringVarP(&patchOpts.component, "component", "c", "", "The component in the bundle to patch")

	patchCmd.AddCommand(patchBundleCmd, patchCompCmd)
	rootCmd.AddCommand(patchCmd)
}

func validateBaseFlags() error {
	if patchOpts.bundlePath == "" {
		return errors.New("a bundle yaml file must be specified")
	}
	if patchOpts.optionsCRs == nil {
		return errors.New("at least one yaml file for the options custom resources")
	}
	return nil
}

func patchBundleAction(cmd *cobra.Command, _ []string) {
	if err := validateBaseFlags(); err != nil {
		exitWithHelp(cmd, err.Error())
	}

	if err := runPatchBundle(patchOpts, &realReaderWriter{}, &localFileSystemReader{}); err != nil {
		log.Exit(err)
	}
}

func runPatchBundle(opts *patchOptions, brw BundleReaderWriter, or OptionsReader) error {
	b, err := readBundle(opts.bundlePath, brw)
	if err != nil {
		return err
	}

	crs, err := readAllOptions(opts.optionsCRs, or)
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
	return brw.WriteBundleFile(opts.output, patched, patchedFilePermissions)
}

func patchCompAction(cmd *cobra.Command, _ []string) {
	if err := validateBaseFlags(); err != nil {
		exitWithHelp(cmd, err.Error())
	}

	if patchOpts.component == "" {
		exitWithHelp(cmd, "the name of the component to patch must be specified.")
	}

	if err := runPatchComponent(patchOpts, &realReaderWriter{}, &localFileSystemReader{}, &localFileSystemWriter{}); err != nil {
		log.Exit(err)
	}
}

func runPatchComponent(opts *patchOptions, brw BundleReaderWriter, or OptionsReader, aw compWriter) error {
	b, err := readBundle(opts.bundlePath, brw)
	if err != nil {
		return err
	}

	crs, err := readAllOptions(opts.optionsCRs, or)
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
	return aw.WriteComponentToFile(patched, opts.output, patchedFilePermissions)
}

// createFinderFn creates an compFinder that operates on the given ClusterBundle.
var createFinderFn = func(b *bpb.ClusterBundle) (compFinder, error) {
	return find.NewBundleFinder(b)
}

// createPatcherFn creates an Patcher that operates on the given ClusterBundle.
var createPatcherFn = func(b *bpb.ClusterBundle) (Patcher, error) {
	return patch.NewPatcherFromBundle(b)
}

func readBundle(path string, brw BundleReaderWriter) (*bpb.ClusterBundle, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	b, err := brw.ReadBundleFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error reading bundle from %s: %v", absPath, err)
	}
	return b, err
}

// readAllOptions reads the options CRs from the given list of yaml files.
// It returns a list of CRs in a map representation, where the CR fields are the map keys.
func readAllOptions(optionsCRs []string, or OptionsReader) ([]map[string]interface{}, error) {
	crs := make([]map[string]interface{}, 0, len(optionsCRs))
	for _, o := range optionsCRs {
		opath, err := filepath.Abs(o)
		if err != nil {
			return nil, err
		}
		bytes, err := or.ReadOptions(opath)
		if err != nil {
			return nil, fmt.Errorf("error reading options from %s: %v", opath, err)
		}
		cr, err := converter.KubeResourceYAMLToMap(bytes)
		if err != nil {
			return nil, err
		}
		crs = append(crs, cr)
	}
	return crs, nil
}
