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

// Command bundlectl is used for modifying bundles.
package main

import (
	"context"
	"flag"
	"os"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands"
)

func main() {
	flag.Lookup("logtostderr").Value.Set("true")

	// Hacky: fixes errors of the form 'ERROR: logging before flag.Parse'.
	// See https://github.com/kubernetes/kubernetes/issues/17162
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse([]string{})

	root := commands.AddCommands(context.Background(), os.Args[1:])
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
