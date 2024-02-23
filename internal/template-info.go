package internal

import (
	"fmt"

	"github.com/null93/vhost/pkg/utils"
	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var templateInfoCmd = &cobra.Command{
	Use:   "info TEMPLATE_NAME",
	Short: "show template details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		if template, err := vhost.LoadTemplate(templateName); err == nil {
			tbl1 := utils.NewTable("Parameter", "Default Value", "Validation", "Custom Validation", "Description")
			tbl1.AddRow(
				"site_name",
				"<injected>",
				"-",
				"-",
				"Name of your site",
			)
			inputSchemaKeys := template.InputSchema.SortedKeys()
			for _, key := range inputSchemaKeys {
				schema := template.InputSchema[key]
				if !schema.ProvisionerOnly {
					tbl1.AddRow(
						key,
						schema.Value,
						schema.Pattern,
						schema.CustomPattern,
						schema.Description,
					)
				}
			}
			tbl1.PrintSeparator()
			fmt.Println("# Template Values")
			fmt.Println("# These values are saved within the checkpoint and can be modified after creation by the user.")
			tbl1.PrintSeparator()
			tbl1.Print()
			tbl1.PrintSeparator()

			tbl2 := utils.NewTable("Parameter", "Default Value", "Validation", "Custom Validation", "Description")
			for _, key := range inputSchemaKeys {
				schema := template.InputSchema[key]
				if schema.ProvisionerOnly {
					tbl2.AddRow(
						key,
						schema.Value,
						schema.Pattern,
						schema.CustomPattern,
						schema.Description,
					)
				}
			}
			fmt.Println("# Provisioner Inputs")
			fmt.Println("# These inputs are used only by the provisioner during creation and are not saved within the checkpoint.")
			tbl2.PrintSeparator()
			tbl2.Print()
			tbl2.PrintSeparator()
		} else {
			ExitWithError(1, fmt.Sprintf("could not find template with name %q", templateName))
		}
	},
}

func init() {
	templateCmd.AddCommand(templateInfoCmd)
}
