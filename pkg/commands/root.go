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
	"context"
	"flag"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/find"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/inline"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/modify"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/validate"
	"github.com/spf13/cobra"
)

// AddCommands adds all subcommands to the root command.
func AddCommands(ctx context.Context, args []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bundlectl",
		Short: "bundlectl is tool for inspecting, validation, and modifying components packages and component sets.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVarP(
		&cmdlib.GlobalOptionsValues.InputFile, "input-file", "f", "", "The path to an input file")

	rootCmd.PersistentFlags().StringVarP(
		&cmdlib.GlobalOptionsValues.InputFormat, "in-format", "", "", "The input file format. One of either 'json' or 'yaml'. "+
			"If an input-file is specified, it is inferred from the file extension. If not specified, it defaults to yaml.")

	rootCmd.PersistentFlags().StringVarP(
		&cmdlib.GlobalOptionsValues.OutputFile, "output-file", "o", "", "The path for any output file")

	rootCmd.PersistentFlags().StringVarP(
		&cmdlib.GlobalOptionsValues.OutputFormat, "format", "", "", "The output file format. One of either 'json' or 'yaml'. "+
			"If an output-file is specified, it is inferred from the file extension. If not specified, it defaults to yaml.")

	rootCmd.PersistentFlags().BoolVarP(
		&cmdlib.GlobalOptionsValues.InlineComponents, "inline-components", "l", true, "Whether to inline the component data files before processing")

	rootCmd.PersistentFlags().BoolVarP(
		&cmdlib.GlobalOptionsValues.InlineObjects, "inline-objects", "", true, "Whether to inline the components' object files")

	filter.AddCommandsTo(ctx, rootCmd)
	find.AddCommandsTo(ctx, rootCmd)
	inline.AddCommandsTo(ctx, rootCmd)
	modify.AddCommandsTo(ctx, rootCmd)
	validate.AddCommandsTo(ctx, rootCmd)

	// This is magic hackery I don't unherdstand but somehow this fixes
	// errrs of the form 'ERROR: logging before flag.Parse'. See more at:
	// https://github.com/kubernetes/kubernetes/issues/17162
	// rootCmd.SetArgs(args)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// pflag.Parse()
	flag.CommandLine.Parse([]string{})

	return rootCmd
}
