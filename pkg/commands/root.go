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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/export"
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
		Use:   "bundler",
		Short: "bundler is tool for inspecting and modifying cluster bundles.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// TODO(kashomon): Should the GlobalOptionsValues be de-globalized? It's
	// certainly possible.
	rootCmd.PersistentFlags().StringVarP(
		&cmdlib.GlobalOptionsValues.BundleFile, "bundle-file", "f", "", "The path to a bundle file")

	rootCmd.PersistentFlags().StringVarP(
		&cmdlib.GlobalOptionsValues.InputFormat, "in-format", "", "yaml", "The input file format. One of either 'json' or 'yaml'")

	rootCmd.PersistentFlags().StringVarP(
		&cmdlib.GlobalOptionsValues.OutputFile, "output-file", "o", "", "The path for any output file")

	rootCmd.PersistentFlags().StringVarP(
		&cmdlib.GlobalOptionsValues.OutputFormat, "format", "", "yaml", "The output file format. One of either 'json' or 'yaml'")

	rootCmd.PersistentFlags().BoolVarP(
		&cmdlib.GlobalOptionsValues.Inline, "inline", "l", true, "Whether to inline the bundle before processing")
	rootCmd.PersistentFlags().BoolVarP(
		&cmdlib.GlobalOptionsValues.TopLayerInlineOnly, "only-inline-top", "", false, "Whether to inline just the top layer of the bundle (node config and component files)")

	export.AddCommandsTo(ctx, rootCmd)
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
