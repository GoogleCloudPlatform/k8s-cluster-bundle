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

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/build"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/export"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/find"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/patch"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/validate"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/version"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/spf13/cobra"
)

// AddCommands adds all subcommands to the root command.
func AddCommands(ctx context.Context, args []string) *cobra.Command {
	cio := &cmdlib.CmdIO{
		FileIO: &files.LocalFileSystemReaderWriter{},
		StdIO:  &cmdlib.RealStdioReaderWriter{},
		ExitIO: &cmdlib.RealExiter{},
	}

	flagset := flag.CommandLine

	return AddCommandsInternal(ctx, cio, flagset, args)
}

// AddCommandsInternal is an internal command that allows for dependency
// injection into sub-commands.
//
// Note: This method is only public to allow for sub-commands to define
// integration tests.
func AddCommandsInternal(ctx context.Context, cio *cmdlib.CmdIO, flagset *flag.FlagSet, args []string) *cobra.Command {
	gopts := &cmdlib.GlobalOptions{}

	rootCmd := &cobra.Command{
		Use:   "bundlectl",
		Short: "bundlectl is tool for inspecting, validation, and modifying components packages and component sets. If a command outputs data, the data is written to STDOUT.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVarP(
		&(gopts.InputFile), "input-file", "f", "", "The path to an input file")

	rootCmd.PersistentFlags().StringVarP(
		&(gopts.InputFormat), "in-format", "", "", "The input file format. One of either 'json' or 'yaml'. "+
			"If an input-file is specified, it is inferred from the file extension. If not specified, it defaults to yaml.")

	rootCmd.PersistentFlags().StringVarP(
		&(gopts.OutputFormat), "format", "", "", "The output file format. One of either 'json' or 'yaml'. "+
			"If not specified, it defaults to yaml.")

	rootCmd.AddCommand(build.GetCommand(ctx, cio.FileIO, cio.StdIO, gopts))
	rootCmd.AddCommand(export.GetCommand(ctx, cio.FileIO, cio.StdIO, gopts))
	rootCmd.AddCommand(filter.GetCommand(ctx, cio.FileIO, cio.StdIO, gopts))
	rootCmd.AddCommand(find.GetCommand(ctx, cio.FileIO, cio.StdIO, gopts))
	rootCmd.AddCommand(patch.GetCommand(ctx, cio.FileIO, cio.StdIO, gopts))
	rootCmd.AddCommand(validate.GetCommand(ctx, cio.FileIO, cio.StdIO, gopts))
	rootCmd.AddCommand(version.GetCommand(cio))

	// This is magic hackery I don't unherdstand but somehow this fixes
	// errrs of the form 'ERROR: logging before flag.Parse'. See more at:
	// https://github.com/kubernetes/kubernetes/issues/17162
	rootCmd.PersistentFlags().AddGoFlagSet(flagset)
	// pflag.Parse()
	flag.CommandLine.Parse([]string{})
	rootCmd.SetArgs(args)

	return rootCmd
}
