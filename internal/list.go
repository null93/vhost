package internal

import (
	"fmt"

	. "github.com/jetrails/jrctl/pkg/output"
	"github.com/jetrails/proposal-nginx/sdk/vhost"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list DOMAIN",
	Short:   "list all configured vhosts",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		tbl := NewTable(Columns{
			"Domain Name",
			"Template Name",
			"Enabled",
		})
		enabledMap := map[bool]string{true: "enabled", false: "disabled"}
		for _, status := range vhost.List() {
			tbl.AddRow(Columns{
				status.VirtualHost.Name,
				status.VirtualHost.Template.Name,
				enabledMap[status.Enabled],
			})
		}
		fmt.Println()
		tbl.PrintTable()
		fmt.Println()
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
