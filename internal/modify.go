package internal

import (
	"fmt"
	"time"

	"github.com/jetrails/proposal-nginx/pkg/vhost"
	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:     "modify SITE_NAME KEY=VALUE...",
	Short:   "modify a site from a template",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: ValidateKeyValueArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]
		templateInput := ParseKeyValueArgs(args[1:])

		if !vhost.SiteExists(siteName) {
			ExitWithError(1, fmt.Sprintf("site %q does not exist", siteName))
		}

		checkPoints, errCheckPoints := vhost.GetCheckPoints(siteName)
		if errCheckPoints != nil {
			ExitWithError(2, errCheckPoints.Error())
		}

		if len(checkPoints) < 1 {
			ExitWithError(3, fmt.Sprintf("site %q has no checkpoints", siteName))
		}

		checkPoint := checkPoints[len(checkPoints)-1]

		checkPoint.Input = MergeInput(checkPoint.Input, templateInput)
		templateOutput, errRender := checkPoint.Template.Render(siteName, checkPoint.Input)
		if errRender != nil {
			ExitWithError(4, errRender.Error())
		}

		if templateOutput.Hash() == checkPoint.Output.Hash() {
			ExitWithError(5, "no changes detected")
		}

		if errDelete := checkPoint.Output.DeleteFiles(true); errDelete != nil {
			ExitWithError(6, errDelete.Error())
		}

		checkPoint.Revision = checkPoint.Revision + 1
		checkPoint.Description = "modified template inputs"
		checkPoint.Timestamp = time.Now()
		checkPoint.Output = templateOutput

		errOutputSave := checkPoint.Output.Save()
		if errOutputSave != nil {
			ExitWithError(7, errOutputSave.Error())
		}

		errSave := checkPoint.Save()
		if errSave != nil {
			ExitWithError(8, errSave.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(modifyCmd)
}
