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

// Package generate contains the generate command.
package generate

import (
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/generate"
	"github.com/spf13/cobra"
)

// GetCommand generate placeholder components
func GetCommand() *cobra.Command {
	var filepath string
	var componentName string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate placeholder components",
		Long:  "Generate placeholder components",
		Run: func(cmd *cobra.Command, _ []string) {
			generate.Create(filepath, componentName)
		},
	}

	cmd.Flags().StringVar(&filepath, "write-to", "", "filepath to create component. By default uses the current working dir.")
	cmd.Flags().StringVar(&componentName, "name", "example-component", "name of component to create")

	return cmd
}
