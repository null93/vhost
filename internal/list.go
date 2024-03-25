package internal

import (
	"fmt"

	"github.com/null93/vhost/pkg/utils"
	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list DOMAIN",
	Short: "list all configured vhosts",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		tbl := utils.NewTable(
			"Site Name",
			"Status",
			"Template",
		)
		sites, errList := vhost.GetSites()
		if errList != nil {
			ExitWithError(1, errList.Error())
		}
		for _, site := range sites {
			template := "<manual>"
			if site.LatestCheckPoint.Template.Exists() {
				template = fmt.Sprintf("%s (%s)", site.LatestCheckPoint.Template.Name, site.LatestCheckPoint.Template.Hash())
			}
			tbl.AddRow(
				site.Name,
				string(site.State),
				template,
			)
		}
		tbl.PrintSeparator()
		tbl.Print()
		tbl.PrintSeparator()
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
