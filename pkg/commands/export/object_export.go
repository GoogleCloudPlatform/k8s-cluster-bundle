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
	log "k8s.io/klog"
)

// objOptions represents options flags for the export objects command.
type objOptions struct{}

func objectAction(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, _ *objOptions, gopt *cmdlib.GlobalOptions) {
	brw := cmdlib.NewBundleReaderWriter(fio, sio)
	if err := runObject(ctx, brw, sio, gopt); err != nil {
		log.Exit(err)
	}
}

func runObject(ctx context.Context, brw cmdlib.BundleReaderWriter, stdio cmdlib.StdioReaderWriter, gopt *cmdlib.GlobalOptions) error {
	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	s, err := converter.NewExporter(bw.AllComponents()...).ObjectsAsSingleYAML()
	if err != nil {
		return err
	}
	_, err = stdio.Write([]byte(s))
	return err
}
