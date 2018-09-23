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

package inline

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/spf13/cobra"
)

// AddCommandsTo adds commands to a root cobra command.
func AddCommandsTo(ctx context.Context, root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "inline",
		Short: "Inline the bundle files",
		Long:  `Inline all the files in the given bundle yaml`,
		Run:   cmdlib.ContextAction(ctx, action),
	}

	// Required flags
	// Note: the path to the bundle must be absolute when running with bazel due
	// to bazel sandboxing
	cmd.Flags().StringVarP(&opts.bundle, "bundle", "b", "",
		"The path to the bundle to inline")

	// Optional flags
	// Note: the path to the output bundle is required when running with bazel,
	// and it must be absolute.
	cmd.Flags().StringVarP(&opts.output, "output", "o", "bundle_inline.yaml",
		"Where to output the inline bundle. By default it outputs to the current working directory")

	root.AddCommand(cmd)
}
