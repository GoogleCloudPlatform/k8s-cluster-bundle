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
	// InputFile is a path to an input file. If a file is not specified, it's
	// assumed the input is provided via STDIN.
	InputFile string

	// InputFormat is the text format for the input. By default, assumes YAML
	// but also can be JSON.
	InputFormat string

	// OutputFormat is the the format for any output. By default, assumes YAML.
	OutputFormat string
}