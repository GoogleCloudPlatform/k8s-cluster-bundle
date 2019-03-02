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

	log "k8s.io/klog"
	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
)

type options struct {}

func findAction(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, cmd *cobra.Command, _ []string, opts *options, goptArgs *cmdlib.GlobalOptions) {
	gopt := goptArgs.Copy()
	brw := cmdlib.NewBundleReaderWriter(fio, sio)
	if err := runFindImages(ctx, brw, gopt); err != nil {
		log.Exitf("error in runFindImages: %v", err)
	}
}

func runFindImages(ctx context.Context, brw cmdlib.BundleReaderWriter, gopt *cmdlib.GlobalOptions) error {
	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading contents: %v", err)
	}

	found := find.NewImageFinder(bw.AllComponents()).AllImages().Flattened()

	return brw.WriteStructuredContents(ctx, found, gopt)
}
