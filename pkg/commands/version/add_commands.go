package version

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/version"
)

func AddCommandsTo(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "List version of bundlectl",
		Long:  "List version of bundlectl",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Println(version.BundlectlVersion)
		},
	}
	root.AddCommand(cmd)
}
