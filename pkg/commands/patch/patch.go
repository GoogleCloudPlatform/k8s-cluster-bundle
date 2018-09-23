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
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/patch"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// patcher provides an interface for patching overlays into a ClusterBundle for
// dependency injection.
type patcher interface {
	PatchBundle(customResources []map[string]interface{}) (*bpb.ClusterBundle, error)
	PatchComponent(component *bpb.ClusterComponent, customResources []map[string]interface{}) (*bpb.ClusterComponent, error)
}

// optionsResourceReader is an interface for reading options custom resources.
type optionsResourceReader interface {
	ReadOptions(filepath string) ([]byte, error)
}

// localFileSystemReader implements the optionsResourceReader interface and reads
// options custom resources from the local filesystem.
type localFileSystemReader struct{}

func (r *localFileSystemReader) ReadOptions(filepath string) ([]byte, error) {
	return ioutil.ReadFile(filepath)
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

func bundleAction(cmd *cobra.Command, _ []string) {
	if err := validateBaseFlags(); err != nil {
		cmdlib.ExitWithHelp(cmd, err.Error())
	}

	if err := runPatchBundle(opts, &cmdlib.RealReaderWriter{}, &localFileSystemReader{}); err != nil {
		log.Exit(err)
	}
}

func runPatchBundle(opts *options, brw cmdlib.BundleReaderWriter, reader optionsResourceReader) error {
	b, err := readBundle(opts.bundlePath, brw)
	if err != nil {
		return err
	}

	crs, err := readAllOptions(opts.optionsCRs, reader)
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
	return brw.WriteBundleFile(opts.output, patched, cmdlib.DefaultFilePermissions)
}

func componentAction(cmd *cobra.Command, _ []string) {
	if err := validateBaseFlags(); err != nil {
		cmdlib.ExitWithHelp(cmd, err.Error())
	}

	if opts.component == "" {
		cmdlib.ExitWithHelp(cmd, "the name of the component to patch must be specified.")
	}

	if err := runPatchComponent(opts, &cmdlib.RealReaderWriter{}, &localFileSystemReader{}, &localFileSystemWriter{}); err != nil {
		log.Exit(err)
	}
}

func runPatchComponent(opts *options, brw cmdlib.BundleReaderWriter, reader optionsResourceReader, aw cmdlib.ComponentWriter) error {
	b, err := readBundle(opts.bundlePath, brw)
	if err != nil {
		return err
	}

	crs, err := readAllOptions(opts.optionsCRs, reader)
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

// createPatcherFn creates an patcher that operates on the given ClusterBundle.
var createPatcherFn = func(b *bpb.ClusterBundle) (patcher, error) {
	return patch.NewPatcherFromBundle(b)
}

// Read a bundle file from disk.
func readBundle(path string, brw cmdlib.BundleReaderWriter) (*bpb.ClusterBundle, error) {
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

// readAllOptions reads the options custome resources from the given list of
// yaml files. It returns a list of CRs in a map representation, where the CR
// fields are the map keys.
func readAllOptions(optionsCRs []string, reader optionsResourceReader) ([]map[string]interface{}, error) {
	crs := make([]map[string]interface{}, 0, len(optionsCRs))
	for _, o := range optionsCRs {
		opath, err := filepath.Abs(o)
		if err != nil {
			return nil, err
		}
		bytes, err := reader.ReadOptions(opath)
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
