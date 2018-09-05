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

package transformer

import (
	"fmt"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
)

// ExportedApp is a ClusterApplication that has been extracted from a ClusterBundle.
type ExportedApp struct {
	// Name represents name of the application.
	Name string

	// Objects represent the cluster objects.
	Objects []*bpb.ClusterObject
}

// AppExporter is a struct that exports cluster apps.
type AppExporter struct {
	finder *find.BundleFinder
}

// NewAppExporter creates a new app exporter.
func NewAppExporter(b *bpb.ClusterBundle) (*AppExporter, error) {
	f, err := find.NewBundleFinder(b)
	if err != nil {
		return nil, err
	}
	return &AppExporter{
		finder: f,
	}, nil
}

// Export extracts the named ClusterApplication from the given bundle. It returns a list of
// ExportedApps.
// - Returns an error if no application by the given appName is found.
// - Returns an error if the desired application in the given bundle is not inlined.
func (e *AppExporter) Export(b *bpb.ClusterBundle, appName string) (*ExportedApp, error) {
	app := e.finder.ClusterApp(appName)
	if app == nil {
		return nil, fmt.Errorf("could not find cluster application named %q", appName)
	}

	objs := app.GetClusterObjects()
	if len(objs) == 0 {
		return nil, fmt.Errorf("no cluster objects found for app %q", appName)
	}

	for _, co := range objs {
		if co.GetInlined() == nil {
			return nil, fmt.Errorf("cluster object %q is not inlined for app %q", co.GetName(), appName)
		}
	}

	return &ExportedApp{
		Name:    appName,
		Objects: objs,
	}, nil
}
