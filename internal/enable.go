package internal

import (
	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable SITE_NAME",
	Short: "enable a site by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]

		if !vhost.SiteExists(siteName) {
			ExitWithError(1, "site does not exist")
		}

		latestCheckPoint, errLatest := vhost.GetLatestCheckPoint(siteName)
		if errLatest != nil {
			ExitWithError(2, errLatest.Error())
		}

		if !latestCheckPoint.Template.Exists() {
			ExitWithError(3, "site is not managed by vhost")
		}

		if err := vhost.EnableSite(siteName); err != nil {
			ExitWithError(4, "failed to enable site")
		}
	},
}

func init() {
	RootCmd.AddCommand(enableCmd)
}
