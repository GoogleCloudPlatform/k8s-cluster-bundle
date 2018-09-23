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

package validate

import (
	"github.com/spf13/cobra"
)

func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a bundle file",
		Long:  `Validate a bundle file to ensure the bundle file follows the bundle schema and doesn't contain errors.`,
		Run:   action,
	}

	// Required flags
	// Note: the path to the bundle must be absolute when running with bazel.
	cmd.Flags().StringVarP(&opts.bundle, "bundle", "b", "",
		"The path to the bundle to validate")

	root.AddCommand(cmd)
}
