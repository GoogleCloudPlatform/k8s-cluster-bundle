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
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

type options struct {
}

var opts = &options{}

func findAction(ctx context.Context, cmd *cobra.Command, _ []string) {
	rw := &files.LocalFileSystemReaderWriter{}
	gopts := cmdlib.GlobalOptionsValues.Copy()
	if err := runFindImages(ctx, opts, rw, gopts); err != nil {
		log.Exitf("error in runFindImages: %v", err)
	}
}

func runFindImages(ctx context.Context, _ *options, rw files.FileReaderWriter, gopt *cmdlib.GlobalOptions) error {
	b, err := cmdlib.ReadBundleContents(ctx, rw, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	found := (&find.ImageFinder{b}).AllImages().Flattened()

	return cmdlib.WriteStructuredContents(ctx, found, rw, gopt)
}
