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

package export

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	bundleoptions "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/spf13/cobra"
	log "k8s.io/klog"
)

// options represents options flags for the export command.
type options struct {
	optionsFiles []string
}

func action(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, cmd *cobra.Command, opts *options, gopt *cmdlib.GlobalOptions) {
	brw := cmdlib.NewBundleReaderWriter(fio, sio)
	if err := run(ctx, opts, brw, fio, sio, gopt); err != nil {
		log.Exit(err)
	}
}

func run(ctx context.Context, o *options, brw cmdlib.BundleReaderWriter, fio files.FileReaderWriter, stdio cmdlib.StdioReaderWriter, gopt *cmdlib.GlobalOptions) error {
	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	var optData bundleoptions.JSONOptions
	if len(o.optionsFiles) > 0 {
		var err error
		optData, err = cmdlib.MergeOptions(ctx, fio, o.optionsFiles)
		if err != nil {
			return err
		}
	}
	objs, err := bw.ExportAsObjects(optData)
	if err != nil {
		return err
	}

	exporter := converter.ObjectExporter{Objects: objs}
	s, err := exporter.ExportAsYAML()
	if err != nil {
		return err
	}
	_, err = stdio.Write([]byte(s))
	return err
}
