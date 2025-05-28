package internal

import (
	"github.com/null93/vhost/pkg/utils"

	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info SITE_NAME [KEY]",
	Short: "show info about a site",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]
		targetKey := ""
		if len(args) > 1 {
			targetKey = args[1]
		}
		if !vhost.SiteExists(siteName) {
			ExitWithError(1, "site does not exist")
		}
		site, errSite := vhost.GetSite(siteName)
		if errSite != nil {
			ExitWithError(2, errSite.Error())
		}
		if !site.LatestCheckPoint.Template.Exists() {
			ExitWithError(3, "site is not managed by vhost")
		}
		tbl := utils.NewTable("Key", "Value")
		if quiet, _ := cmd.Flags().GetBool("quiet"); quiet {
			tbl.SetQuietCols("Value")
		}
		if targetKey == "" || targetKey == "state" {
			tbl.AddRow("state:", string(site.State))
		}
		if targetKey == "" || targetKey == "template_hash" {
			tbl.AddRow("template_hash:", site.LatestCheckPoint.Template.Hash())
		}
		if targetKey == "" || targetKey == "input_hash" {
			tbl.AddRow("input_hash:", site.LatestCheckPoint.Input.Hash())
		}
		if targetKey == "" || targetKey == "output_hash" {
			tbl.AddRow("output_hash:", site.LatestCheckPoint.Output.Hash())
		}
		for key, value := range site.LatestCheckPoint.Input {
			if targetKey == "" || targetKey == key {
				tbl.AddRow(key+":", value)
			}
		}
		tbl.PrintSeparator()
		tbl.Print()
		tbl.PrintSeparator()
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolP("quiet", "q", false, "minimal output")
}
