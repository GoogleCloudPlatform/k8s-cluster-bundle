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
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/spf13/cobra"
)

// AddCommandsTo adds commands to a root cobra command.
func AddCommandsTo(ctx context.Context, root *cobra.Command) {
	// cmd is the parent patch command, and is unrunnable by itself.
	// Patch subcommands should be added to it.
	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Apply patches to a bundle or part of a bundle",
		Long:  "Apply patches for bundle customization provided the given options custom resources. See subcommands for patch usage.",
	}

	cmd.PersistentFlags().StringSliceVarP(&opts.optionsCRs, "options-resources", "p", nil,
		"The yaml files containing the options custom resource(s) (comma separated)")

	bundleCmd := &cobra.Command{
		Use:   "bundle",
		Short: "Apply patches to the entire bundle",
		Long:  "Apply all the patches found in a bundle to customize it with the given options custom resources",
		Run:   cmdlib.ContextAction(ctx, bundleAction),
	}

	componentCmd := &cobra.Command{
		Use:   "component",
		Short: "Apply patches to a component in the bundle",
		Long:  "Apply all the patches found in the given component in the bundle to customize it with the given options custom resources",
		Run:   cmdlib.ContextAction(ctx, componentAction),
	}

	// Required patch component flags
	componentCmd.Flags().StringVarP(&opts.component, "component", "c", "", "The component in the bundle to patch")

	cmd.AddCommand(bundleCmd, componentCmd)
	root.AddCommand(cmd)
}
