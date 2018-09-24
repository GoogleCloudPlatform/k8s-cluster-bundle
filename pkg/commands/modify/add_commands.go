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

package modify

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/spf13/cobra"
)

// AddCommandsTo adds commands to a root cobra command.
func AddCommandsTo(ctx context.Context, root *cobra.Command) {
	// cmd is the parent image command, and is unrunnable by itself.
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Modify objects inside the bundle",
		Long:  "Provides functionality for searching for or modifying container or node images. See subcommands for patch usage.",
	}

	imagesCmd := &cobra.Command{
		Use:   "images",
		Short: "Find images in the bundle",
		Long:  "Apply all the patches found in a bundle to customize it with the given options custom resources",
		Run:   cmdlib.ContextAction(ctx, modifyImagesAction),
	}

	// Required flags
	// Note: the path to the bundle must be absolute when running with bazel.
	imagesCmd.Flags().StringVarP(&opts.findReplacePairs, "find-replace-pairs", "", "",
		"Pairs of strings to preform find and replace. Should have the form \"find,replace;find,replace\"")

	cmd.AddCommand(imagesCmd)
	root.AddCommand(cmd)
}
