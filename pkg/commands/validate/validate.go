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

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validation"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// options contain options flags for the bundle validation command.
type options struct {
	// bundle is a filepath to a bundle
	bundle string
}

// opts is a global options flags instance for reference via the cobra command
// installation.
var opts = &options{}

// Action is the cobra command action for bundle validation.
func action(ctx context.Context, cmd *cobra.Command, _ []string) {
	if opts.bundle == "" {
		cmdlib.ExitWithHelp(cmd, "Please provide yaml file for bundle.")
	}
	if err := runValidate(ctx, opts, converter.NewFileSystemBundleReaderWriter()); err != nil {
		log.Exit(err)
	}
}

type bundleValidator interface {
	Validate() []error
}

// createValidatorFn creates BundleValidator that works with the given current
// working directory and allows for dependency injection.
var createValidatorFn = func(b *bpb.ClusterBundle) bundleValidator {
	return validation.NewBundleValidator(b)
}

func runValidate(ctx context.Context, opts *options, brw *converter.BundleReaderWriter) error {
	b, err := brw.ReadBundleFile(ctx, opts.bundle)
	if err != nil {
		return err
	}

	val := createValidatorFn(b)
	if errs := val.Validate(); len(errs) > 0 {
		return fmt.Errorf("there were one or more errors found while validating the bundle:\n%v", validation.JoinErrors(errs))
	}
	log.Info("No errors found")
	return nil
}
