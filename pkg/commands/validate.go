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

package commands

import (
	"fmt"
	"path/filepath"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validation"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

type validateOptions struct {
	bundle string
}

var validateOpts = &validateOptions{}

func addValidateCommand() {
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a bundle file",
		Long:  `Validate a bundle file to ensure the bundle file follows the bundle schema and doesn't contain errors.`,
		Run:   validateAction,
	}

	// Required flags
	// Note: the path to the bundle must be absolute when running with bazel.
	validateCmd.Flags().StringVarP(&validateOpts.bundle, "bundle", "b", "",
		"The path to the bundle to validate")

	rootCmd.AddCommand(validateCmd)
}

func validateAction(cmd *cobra.Command, _ []string) {
	if validateOpts.bundle == "" {
		exitWithHelp(cmd, "Please provide yaml file for bundle.")
	}
	if err := runValidate(validateOpts, &realReaderWriter{}); err != nil {
		log.Exit(err)
	}
}

type bundleValidator interface {
	Validate() []error
}

// createValidatorFn creates BundleValidator that works with the given current
// working directory.
var createValidatorFn = func(b *bpb.ClusterBundle) bundleValidator {
	return validation.NewBundleValidator(b)
}

func runValidate(opts *validateOptions, brw BundleReaderWriter) error {
	path, err := filepath.Abs(opts.bundle)
	if err != nil {
		return err
	}

	b, err := brw.ReadBundleFile(path)
	if err != nil {
		return err
	}

	validator := createValidatorFn(b)
	if errs := validator.Validate(); len(errs) > 0 {
		return fmt.Errorf("there were one or more errors found while validating the bundle:\n%v", validation.JoinErrors(errs))
	}
	log.Info("No errors found")
	return nil
}
