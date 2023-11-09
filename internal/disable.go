package internal

import (
	"fmt"

	"github.com/jetrails/proposal-nginx/pkg/vhost"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable SITE_NAME",
	Short: "disable a site by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]

		if !vhost.SiteExists(siteName) {
			ExitWithError(1, fmt.Sprintf("site %q does not exist", siteName))
		}

		if err := vhost.DisableSite(siteName); err != nil {
			ExitWithError(2, "failed to disable site")
		}
	},
}

func init() {
	RootCmd.AddCommand(disableCmd)
}
