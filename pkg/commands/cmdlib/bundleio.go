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

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
	log "github.com/golang/glog"
)

func ReadBundleContents(ctx context.Context, rw core.FileReaderWriter, g *GlobalOptions) (*bpb.ClusterBundle, error) {
	var bytes []byte
	var err error
	if g.BundleFile != "" {
		log.Infof("Reading bundle file %v", g.BundleFile)
		bytes, err = rw.ReadFile(ctx, g.BundleFile)
		if err != nil {
			return nil, err
		}
	} else {
		log.Info("No bundle file, reading from stdin")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			bytes = append(bytes, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	b, err := convertBundle(ctx, bytes, rw, g)
	if err != nil {
		return nil, err
	}

	// For now, we can only inline bundle files because we need the path context.
	if g.Inline && g.BundleFile != "" {
		return inlineBundle(ctx, b, g.BundleFile, rw)
	}
	return b, nil
}

// convertBundle converts a bundle from raw bytes into a proto.
func convertBundle(ctx context.Context, bytes []byte, rw core.FileReaderWriter, g *GlobalOptions) (*bpb.ClusterBundle, error) {
	inFormat := strings.ToLower(g.InputFormat)
	if inFormat == "yaml" {
		bi, err := converter.Bundle.YAMLToProto(bytes)
		if err != nil {
			return nil, fmt.Errorf("error converting bundle from yaml to proto: %v", err)
		}
		return converter.ToBundle(bi), nil
	} else if inFormat == "json" {
		bi, err := converter.Bundle.JSONToProto(bytes)
		if err != nil {
			return nil, fmt.Errorf("error converting bundle from json to proto: %v", err)
		}
		return converter.ToBundle(bi), nil
	}
	return nil, fmt.Errorf("unknown input format: %q", inFormat)
}

// inlineBundle inlines a cluster bundle before processing
func inlineBundle(ctx context.Context, b *bpb.ClusterBundle, path string, rw core.FileReaderWriter) (*bpb.ClusterBundle, error) {
	inliner := &transformer.Inliner{
		&core.LocalFilePBReader{
			WorkingDir: filepath.Dir(path),
			Rdr:        rw,
		},
	}
	return inliner.Inline(ctx, b)
}

// WriteContents writes some bytes to disk or stdout. if outPath is empty, write to stdout instdea.
func WriteContents(ctx context.Context, outPath string, bytes []byte, rw core.FileReaderWriter) error {
	if outPath == "" {
		_, err := os.Stdout.Write(bytes)
		if err != nil {
			return fmt.Errorf("error writing content %q to stdout", string(bytes))
		}
		return nil
	}
	log.Infof("Writing file to %q", outPath)
	return rw.WriteFile(ctx, outPath, bytes, DefaultFilePermissions)
}
