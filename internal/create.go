package internal

import (
	"fmt"

	"github.com/jetrails/proposal-nginx/pkg/vhost"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:     "create TEMPLATE_NAME SITE_NAME KEY=VALUE...",
	Short:   "create a site from a template",
	Args:    cobra.MinimumNArgs(2),
	PreRunE: ValidateKeyValueArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		siteName := args[1]
		templateInput := ParseKeyValueArgs(args[2:])

		if vhost.SiteExists(siteName) {
			ExitWithError(1, fmt.Sprintf("site %q already exists", siteName))
		}

		template, errTemplate := vhost.LoadTemplate(templateName)
		if errTemplate != nil {
			ExitWithError(2, errTemplate.Error())
		}

		checkPoint, errCheckPoint := vhost.NewCheckPoint(siteName, template, templateInput)
		if errCheckPoint != nil {
			ExitWithError(3, errCheckPoint.Error())
		}

		errOutputSave := checkPoint.Output.Save()
		if errOutputSave != nil {
			ExitWithError(4, errOutputSave.Error())
		}

		if errProvisioner := checkPoint.Template.RunProvisioner(siteName, templateInput); errProvisioner == nil {
			checkPoint.Description = "initial"
			errSave := checkPoint.Save()
			if errSave != nil {
				ExitWithError(5, errSave.Error())
			}
		} else {
			checkPoint.Output.DeleteFiles(true)
			ExitWithError(6, errProvisioner.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
}
