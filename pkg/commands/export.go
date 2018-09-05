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
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// exportedAppFilePermissions is the permission used for the exported app output files (rw-r--r--).
const exportedAppFilePermissions = os.FileMode(0644)

// Exporter provides an interface for exporting ClusterApplications from a ClusterBundle.
type Exporter interface {
	Export(b *bpb.ClusterBundle, appName string) (*transformer.ExportedApp, error)
}

type exportOptions struct {
	bundlePath string
	apps       []string
	outputDir  string
}

var exportOpts = &exportOptions{}

func addExportCommand() {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export application(s) from a bundle",
		Long:  `Export the specified application(s) from the given bundle yaml into their own yamls`,
		Run:   exportAction,
	}

	// Required flags
	// Note: the path to the bundle must be absolute when running with bazel.
	exportCmd.Flags().StringVarP(&exportOpts.bundlePath, "bundle", "b", "", "The path to the bundle to inline")
	exportCmd.Flags().StringSliceVarP(&exportOpts.apps, "apps", "a", nil, "The app(s) to extract from the bundle (comma separated)")

	// Optional flags
	// Note: the output directory path is required when running with bazel, and it must be absolute.
	exportCmd.Flags().StringVarP(&exportOpts.outputDir, "output", "o", "",
		"Where to write the extracted apps. By default writes to current working directory")

	rootCmd.AddCommand(exportCmd)
}

func exportAction(cmd *cobra.Command, _ []string) {
	if exportOpts.bundlePath == "" {
		exitWithHelp(cmd, "Please provide yaml file for bundle.")
	}
	if exportOpts.apps == nil {
		exitWithHelp(cmd, "Please provide at least one app to extract.")
	}
	if err := runExport(exportOpts, &realReaderWriter{}, &localFileSystemWriter{}); err != nil {
		log.Exit(err)
	}
}

// createExporterFn creates an Exporter that operates on the given ClusterBundle.
var createExporterFn = func(b *bpb.ClusterBundle) (Exporter, error) {
	return transformer.NewAppExporter(b)
}

func runExport(opts *exportOptions, brw BundleReaderWriter, aw appWriter) error {
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

	for _, app := range opts.apps {
		ea, err := exporter.Export(b, app)
		if err != nil {
			return err
		}
		// If a write fails, just return the error and the user can rerun the command and rewrite
		// any files that may have been written or partially written.
		if err := writeApp(ea, opts.outputDir, exportedAppFilePermissions, aw); err != nil {
			return err
		}
		log.Infof("Wrote file %q", ea.Name)
	}
	return nil
}

// writeApp writes an exported application to the given directory
// and names the file by the app name.
func writeApp(ea *transformer.ExportedApp, dir string, permissions os.FileMode, aw appWriter) error {
	path := fmt.Sprintf("%s/%s.yaml", filepath.Clean(dir), ea.Name)
	app := bpb.ClusterApplication{
		Name:           ea.Name,
		ClusterObjects: ea.Objects,
	}
	return aw.WriteAppToFile(&app, path, permissions)
}
