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

package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// exportedCompFilePermissions is the permission used for the exported
// component output files (rw-r--r--).
const exportedCompFilePermissions = os.FileMode(0644)

// Exporter provides an interface for exporting ClusterComponents from a ClusterBundle.
type Exporter interface {
	Export(b *bpb.ClusterBundle, compName string) (*transformer.ExportedComponent, error)
}

type exportOptions struct {
	bundlePath string
	components []string
	outputDir  string
}

var exportOpts = &exportOptions{}

func addExportCommand() {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export component(s) from a bundle",
		Long:  `Export the specified components(s) from the given bundle yaml into their own yamls`,
		Run:   exportAction,
	}

	// Required flags
	// Note: the path to the bundle must be absolute when running with bazel.
	exportCmd.Flags().StringVarP(&exportOpts.bundlePath, "bundle", "b", "", "The path to the bundle to inline")
	exportCmd.Flags().StringSliceVarP(&exportOpts.components, "components", "c", nil, "The components(s) to extract from the bundle (comma separated)")

	// Optional flags
	// Note: the output directory path is required when running with bazel, and it must be absolute.
	exportCmd.Flags().StringVarP(&exportOpts.outputDir, "output", "o", "",
		"Where to write the extracted components. By default writes to current working directory")

	rootCmd.AddCommand(exportCmd)
}

func exportAction(cmd *cobra.Command, _ []string) {
	if exportOpts.bundlePath == "" {
		exitWithHelp(cmd, "Please provide yaml file for bundle.")
	}
	if exportOpts.components == nil {
		exitWithHelp(cmd, "Please provide at least one component to extract.")
	}
	if err := runExport(exportOpts, &realReaderWriter{}, &localFileSystemWriter{}); err != nil {
		log.Exit(err)
	}
}

// createExporterFn creates an Exporter that operates on the given ClusterBundle.
var createExporterFn = func(b *bpb.ClusterBundle) (Exporter, error) {
	return transformer.NewComponentExporter(b)
}

func runExport(opts *exportOptions, brw BundleReaderWriter, aw compWriter) error {
	path, err := filepath.Abs(opts.bundlePath)
	if err != nil {
		return err
	}

	b, err := brw.ReadBundleFile(path)
	if err != nil {
		return err
	}

	exporter, err := createExporterFn(b)
	if err != nil {
		return err
	}

	for _, comp := range opts.components {
		ea, err := exporter.Export(b, comp)
		if err != nil {
			return err
		}
		// If a write fails, just return the error and the user can rerun the command and rewrite
		// any files that may have been written or partially written.
		if err := writeComp(ea, opts.outputDir, exportedCompFilePermissions, aw); err != nil {
			return err
		}
		log.Infof("Wrote file %q", ea.Name)
	}
	return nil
}

// writeComp writes an exported component to the given directory
// and names the file by the component name.
func writeComp(ea *transformer.ExportedComponent, dir string, permissions os.FileMode, aw compWriter) error {
	path := fmt.Sprintf("%s/%s.yaml", filepath.Clean(dir), ea.Name)
	comp := bpb.ClusterComponent{
		Name:           ea.Name,
		ClusterObjects: ea.Objects,
	}
	return aw.WriteComponentToFile(&comp, path, permissions)
}
