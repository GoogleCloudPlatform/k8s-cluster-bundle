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

package validate

import (
	"context"
	"fmt"

	log "github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validate"
)

// options contain options flags for the bundle validation command.
type options struct {
}

// opts is a global options flags instance for reference via the cobra command
// installation.
var opts = &options{}

// Action is the cobra command action for bundle validation.
func action(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, cmd *cobra.Command, _ []string) {
	gopt := cmdlib.GlobalOptionsValues.Copy()
	brw := cmdlib.NewBundleReaderWriter(fio, sio)
	if err := runValidate(ctx, opts, brw, gopt); err != nil {
		log.Exit(err)
	}
}

func runValidate(ctx context.Context, opts *options, brw cmdlib.BundleReaderWriter, gopt *cmdlib.GlobalOptions) error {
	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading contents: %v", err)
	}

	bundleType := bw.Kind()
	if bundleType == "Component" {
		errs := validate.Component(bw.Component())
		if len(errs) > 0 {
			return fmt.Errorf("there were one or more errors found while validating the bundle:\n%v", errs.ToAggregate())
		}
	} else { //@todo add validation for BundleBuilder, Bundle, ComponentBuilder here.
		return fmt.Errorf("Kind %q not yet supported", bundleType)
	}

	log.Info("No errors found")
	return nil
}
