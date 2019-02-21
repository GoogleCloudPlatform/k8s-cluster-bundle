package version

import (
	"github.com/spf13/cobra"
	"fmt"
)

func AddCommandsTo(root *cobra.Command, bundlectlVersion string) {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "List version of bundlectl",
		Long:  "List version of bundlectl",
		Run:  func(cmd *cobra.Command, _ []string) {
			fmt.Println(bundlectlVersion)
		},
	}
	root.AddCommand(cmd)
}
