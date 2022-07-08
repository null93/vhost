package internal

import (
	"fmt"

	. "github.com/jetrails/jrctl/pkg/output"
	"github.com/jetrails/proposal-nginx/sdk/vhost"
	"github.com/spf13/cobra"
)

var templateListCmd = &cobra.Command{
	Use:     "list",
	Short:   "list available templates",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		tbl := NewTable(Columns{
			"Template Name",
			"Hash",
		})
		for _, template := range vhost.ListTemplates() {
			hash := template.Hash ()
			tbl.AddRow(Columns{
				template.Name,
				hash[:8] + "..." + hash[len(hash)-8:],
			})
		}
		fmt.Println()
		tbl.PrintTable()
		fmt.Println()
	},
}

func init() {
	templateCmd.AddCommand(templateListCmd)
}
