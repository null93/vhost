package internal

import (
	"github.com/null93/vhost/pkg/utils"

	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info SITE_NAME",
	Short: "show info about a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]
		if !vhost.SiteExists(siteName) {
			ExitWithError(1, "site does not exist")
		}
		site, errSite := vhost.GetSite(siteName)
		if errSite != nil {
			ExitWithError(2, errSite.Error())
		}
		tbl := utils.NewTable("Key", "Value")
		tbl.AddRow("state:", string(site.State))
		tbl.AddRow("template_hash:", site.LatestCheckPoint.Template.Hash())
		tbl.AddRow("input_hash:", site.LatestCheckPoint.Input.Hash())
		tbl.AddRow("output_hash:", site.LatestCheckPoint.Output.Hash())
		for key, value := range site.LatestCheckPoint.Input {
			tbl.AddRow(key+":", value)
		}
		tbl.PrintSeparator()
		tbl.Print()
		tbl.PrintSeparator()
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
}
