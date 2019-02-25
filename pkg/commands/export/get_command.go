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

// Package export contains commands for exporting components and objects.
package export

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/spf13/cobra"
)

// GetCommand returns the export command.
func GetCommand(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, gopts *cmdlib.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Exports components or objects",
		Long:  `Exports components or objects. See subcommands for usage`,
	}

	objOpts := &objOptions{}
	objectCmd := &cobra.Command{
		Use:   "objects",
		Short: "Exports all of the objects",
		Long:  `Exports all objects to STDOUT as YAML delimited by ---`,
		Run: func(cmd *cobra.Command, args []string) {
			objectAction(ctx, fio, sio, objOpts, gopts)
		},
	}

	compOpts := &compOptions{}
	componentCmd := &cobra.Command{
		Use:   "components",
		Short: "Exports all of the components",
		Long:  `Exports all of the components, to STDOUT or to files.`,
		Run: func(cmd *cobra.Command, args []string) {
			componentAction(ctx, fio, sio, compOpts, gopts)
		},
	}

	componentCmd.Flags().StringVar(&compOpts.writeTo, "write-to", "",
		"Directory to write the components to. If not set, writes to STDOUT, delimited by ---")
	componentCmd.Flags().BoolVar(&compOpts.overwrite, "overwrite", false,
		"Whether to overwrite files. By default, if a file already exists, an error will be produced.")

	componentCmd.Flags().BoolVar(&compOpts.exportSet, "export-set", false,
		"Whether to export a component set when exporting components.")
	componentCmd.Flags().StringVar(&compOpts.setName, "setName", "component-set",
		"name for the component set, if not specified by a bundle")
	componentCmd.Flags().StringVar(&compOpts.setVersion, "setVersion", "0.1.0",
		"version for the component set, if not specified by a bundle")

	cmd.AddCommand(objectCmd)
	cmd.AddCommand(componentCmd)
	return cmd
}
