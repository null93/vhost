package internal

import (
	"fmt"

	"github.com/null93/vhost/pkg/vhost"
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

		latestCheckPoint, errLatest := vhost.GetLatestCheckPoint(siteName)
		if errLatest != nil {
			ExitWithError(2, errLatest.Error())
		}

		if !latestCheckPoint.Template.Exists() {
			ExitWithError(3, "site is not managed by vhost")
		}

		if err := vhost.DisableSite(siteName); err != nil {
			ExitWithError(4, "failed to disable site")
		}
	},
}

func init() {
	RootCmd.AddCommand(disableCmd)
}
