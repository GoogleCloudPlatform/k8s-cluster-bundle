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

package cmdlib

// GlobalOptions are options that apply to all commands
type GlobalOptions struct {
	// A path to a bundle file.
	BundleFile string

	// Format for the input. By default, assumes YAML.
	InputFormat string

	// Path for an output file.
	OutputFile string

	// The format for any output. By default, assumes YAML.
	OutputFormat string

	// Whether to inline the bundle before doing any processing
	Inline bool
}

// GlobalOptionsValues is a global tracker for global options (for command line
// executions only).
var GlobalOptionsValues = &GlobalOptions{}

func (g *GlobalOptions) Copy() *GlobalOptions {
	return &GlobalOptions{
		BundleFile:   g.BundleFile,
		InputFormat:  g.InputFormat,
		OutputFile:   g.OutputFile,
		OutputFormat: g.OutputFormat,
		Inline:       g.Inline,
	}
}
