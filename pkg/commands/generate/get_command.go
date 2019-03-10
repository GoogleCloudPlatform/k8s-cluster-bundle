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
package generate

import (
	"os"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/generate"
)

// GetCommand generate placeholder components
func GetCommand() *cobra.Command {
	var filepath string
	var includeTests bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate placeholder components",
		Long:  "Generate placeholder components",
		Run: func(cmd *cobra.Command, _ []string) {
			if filepath == "" {
				os.Stderr.WriteString("Filepath not specified")
				os.Exit(1)
			}
			fmt.Println("include tests: ", includeTests)
			generate.GenerateComponent(filepath, includeTests)
		},
	}

	cmd.Flags().StringVar(&filepath, "path", "", "filepath to create component")
	cmd.Flags().BoolVarP(&includeTests, "include-test-suite", "", false, "include test suite or not")


	return cmd
}