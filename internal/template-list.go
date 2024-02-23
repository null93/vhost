package internal

import (
	"github.com/null93/vhost/pkg/utils"
	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "list available templates",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		tbl := utils.NewTable(
			"Template Name",
			"Template Hash",
		)
		for _, template := range vhost.GetTemplates() {
			tbl.AddRow(
				template.Name,
				template.Hash(),
			)
		}
		tbl.PrintSeparator()
		tbl.Print()
		tbl.PrintSeparator()
	},
}

func init() {
	templateCmd.AddCommand(templateListCmd)
}
