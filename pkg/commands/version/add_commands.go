package version

import (
	"context"
	"github.com/spf13/cobra"
	"fmt"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
)

func AddCommandsTo(ctx context.Context, root *cobra.Command, bundlectlVersion string) {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "List version of bundlectl",
		Long:  "List version of bundlectl",
		Run:   cmdlib.ContextAction(ctx, func(ctx context.Context, cmd *cobra.Command, _ []string) {
			fmt.Println(bundlectlVersion)
		}),
	}
	root.AddCommand(cmd)
}
