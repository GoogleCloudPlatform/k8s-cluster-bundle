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

package find

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
	"github.com/ghodss/yaml"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// options represents options flags for the images command.
type options struct {
}

// opts is a global options flags instance for reference via the cobra command
// installation.
var opts = &options{}

func findAction(ctx context.Context, cmd *cobra.Command, _ []string) {
	rw := &core.LocalFileSystemReaderWriter{}
	gopts := cmdlib.GlobalOptionsValues
	if err := runFindImages(ctx, opts, rw, gopts); err != nil {
		log.Exitf("error in runFindImages: %v", err)
	}
}

func runFindImages(ctx context.Context, opts *options, rw core.FileReaderWriter, gopt *cmdlib.GlobalOptions) error {
	b, err := cmdlib.ReadBundleContents(ctx, rw, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	finder := find.ImageFinder{b}
	found := finder.AllImages().Flattened()

	fmt.Printf("%+v\n", gopt)

	var bytes []byte
	if gopt.OutputFormat == "yaml" {
		bytes, err = yaml.Marshal(found)
		if err != nil {
			return fmt.Errorf("error marshalling yaml: %v", err)
		}
	} else if gopt.OutputFormat == "json" {
		bytes, err = json.Marshal(found)
		if err != nil {
			return fmt.Errorf("error marshalling json: %v", err)
		}
	}

	outPath := gopt.OutputFile
	err = cmdlib.WriteContents(ctx, outPath, bytes, rw)
	if err != nil {
		return fmt.Errorf("error writing contents: %v", err)
		return err
	}
	return nil
}
