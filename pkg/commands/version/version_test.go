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

package version_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdrunner"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdtest"
)

var (
	// TODO(kashomon): Use a real SemVer library to test this.
	numPattern     = `([1-9]\d*|0)`
	versionPattern = regexp.MustCompile(fmt.Sprintf(`^%s\.%s\.%s$`, numPattern, numPattern, numPattern))
)

func TestVersion(t *testing.T) {
	fakeio := cmdtest.NewFakeCmdIO()

	if err := cmdrunner.ExecuteCommand(fakeio, []string{
		"version",
	}); err != nil {
		t.Fatal(err)
	}

	version := string(fakeio.StdIO.WriteBytes)
	if !versionPattern.MatchString(version) {
		t.Errorf("Got version %s, but expected it to have the form X.Y.Z", version)
	}
}
