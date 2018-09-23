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

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/export"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/inline"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/patch"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/validate"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "bundler",
		Short: "bundler is tool for inspecting and modifying cluster bundles.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

// AddCommands adds all subcommands to the root command.
// Each subcommand should implement an addCommand function to be called here.
// This allows context to be passed down to subcommands.
func AddCommands(ctx context.Context) error {
	export.AddCommandsTo(ctx, rootCmd)
	inline.AddCommandsTo(ctx, rootCmd)
	patch.AddCommandsTo(ctx, rootCmd)
	validate.AddCommandsTo(ctx, rootCmd)
	return nil
}

// Execute invokes the root command and any subcommands that were called.
func Execute() error {
	return rootCmd.Execute()
}
