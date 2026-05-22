package internal

import (
	"fmt"

	"github.com/null93/vhost/pkg/vhost"
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
		become, _ := cmd.Flags().GetString("become")
		skipProvisioner, _ := cmd.Flags().GetBool("skip-provisioner")

		if vhost.SiteExists(siteName) {
			ExitWithError(1, fmt.Sprintf("site %q already exists", siteName))
		}

		template, errTemplate := vhost.LoadTemplate(templateName)
		if errTemplate != nil {
			ExitWithError(2, errTemplate.Error())
		}

		var checkPoint vhost.CheckPoint
		var errCheckPoint error
		if skipProvisioner {
			checkPoint, errCheckPoint = vhost.NewCheckPointWithoutProvisioner(siteName, template, templateInput)
		} else {
			checkPoint, errCheckPoint = vhost.NewCheckPoint(siteName, template, templateInput)
		}
		if errCheckPoint != nil {
			ExitWithError(3, errCheckPoint.Error())
		}

		errOutputSave := checkPoint.Output.Save()
		if errOutputSave != nil {
			ExitWithError(4, errOutputSave.Error())
		}

		if skipProvisioner {
			checkPoint.Description = "initial (provisioner skipped)"
			if errSave := checkPoint.Save(); errSave != nil {
				ExitWithError(5, errSave.Error())
			}
			return
		}

		if errProvisioner := checkPoint.Template.RunProvisioner(siteName, templateInput, become); errProvisioner == nil {
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
	createCmd.Flags().StringP("become", "b", "", "run provisioner as another user")
	createCmd.Flags().Bool("skip-provisioner", false, "create the site without running the provisioner (skips validation of provisioner_only inputs)")
}
