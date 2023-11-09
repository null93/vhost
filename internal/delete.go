package internal

import (
	"fmt"

	"github.com/jetrails/proposal-nginx/pkg/vhost"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete SITE_NAME",
	Short: "delete a site by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]

		if !vhost.SiteExists(siteName) {
			ExitWithError(1, fmt.Sprintf("site %q does not exist", siteName))
		}

		if errDelete := vhost.DeleteSite(siteName, true); errDelete != nil {
			ExitWithError(2, errDelete.Error())
		}

		if errPurge := vhost.PurgeCheckPoints(siteName, true); errPurge != nil {
			ExitWithError(3, errPurge.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
