package internal

import (
	"fmt"
	"os"

	. "github.com/jetrails/jrctl/pkg/output"
	"github.com/jetrails/proposal-nginx/sdk/vhost"
	"github.com/spf13/cobra"
)

func dereference (ptr * string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

var templateInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "show template details",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		if template, err := vhost.LoadTemplate (templateName); err == nil {
			tbl := NewTable(Columns{"Parameter","Default Value", "Validation", "Description"})
			for key, schema := range template.Schema {
				tbl.AddRow (Columns{
					key,
					dereference(schema.Value),
					dereference(schema.Pattern),
					dereference(schema.Description),
				})
			}
			fmt.Println()
			tbl.PrintTable()
			fmt.Println()
		} else {
			fmt.Printf ("\nError: could not find template with name %q\n\n", templateName)
			os.Exit (1)
		}
	},
}

func init() {
	templateCmd.AddCommand(templateInfoCmd)
}
