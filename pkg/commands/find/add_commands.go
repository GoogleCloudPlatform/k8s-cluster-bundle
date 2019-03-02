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

package find

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/spf13/cobra"
)

// AddCommandsTo adds commands to a root cobra command.
func AddCommandsTo(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, root *cobra.Command, gopts *cmdlib.GlobalOptions) {
	opts := &options{}

	// cmd is the parent image command, and is unrunnable by itself.
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Search for objects inside the bundle",
		Long:  "Provides functionality for searching through cluster bundles. See subcommands usage.",
	}

	imagesCmd := &cobra.Command{
		Use:   "images",
		Short: "Find images in the bundle",
		Long:  "Apply all the patches found in a bundle to customize it with the given options custom resources",
		Run: func(cmd *cobra.Command, args[] string) {
			findAction(ctx, fio, sio, cmd, args, opts, gopts)
		},
	}

	cmd.AddCommand(imagesCmd)
	root.AddCommand(cmd)
}
