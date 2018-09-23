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

package export

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/spf13/cobra"
)

// AddCommandsTo adds commands to a root cobra command.
func AddCommandsTo(ctx context.Context, root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export component(s) from a bundle",
		Long:  `Export the specified components(s) from the given bundle yaml into their own yamls`,
		Run:   cmdlib.ContextAction(ctx, action),
	}

	// Required flags
	// Note: the path to the bundle must be absolute when running with bazel.
	cmd.Flags().StringVarP(&opts.bundlePath, "bundle", "b", "", "The path to the bundle to inline")
	cmd.Flags().StringSliceVarP(&opts.components, "components", "c", nil, "The components(s) to extract from the bundle (comma separated)")

	// Optional flags
	// Note: the output directory path is required when running with bazel, and it must be absolute.
	cmd.Flags().StringVarP(&opts.outputDir, "output", "o", "",
		"Where to write the extracted components. By default writes to current working directory")

	root.AddCommand(cmd)
}
