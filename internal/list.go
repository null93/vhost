package internal

import (
	"github.com/jetrails/proposal-nginx/pkg/utils"
	"github.com/jetrails/proposal-nginx/pkg/vhost"
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
			tbl.AddRow(
				site.Name,
				string(site.State),
				site.LatestCheckPoint.Template.Name+" ("+site.LatestCheckPoint.Template.Hash()+")",
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
