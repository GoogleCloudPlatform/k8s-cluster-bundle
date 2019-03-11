// Copyright 2019 Google LLC
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

// Package version contains the version command.
package version

import (
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/version"
	"github.com/spf13/cobra"
)

// GetCommand prints the version of bundlectl.
func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "List version of bundlectl",
		Long:  "List version of bundlectl",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Println(version.BundlectlVersion)
		},
	}
	return cmd
}
