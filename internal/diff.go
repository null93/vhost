package internal

import (
	"fmt"
	"io/ioutil"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:     "diff SITE_NAME",
	Short:   "diff site compared to the template that was used to render it",
	Args:    cobra.ExactArgs(1),
	PreRunE: ValidateKeyValueArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]

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

		if !checkPoint.Template.Exists() {
			ExitWithError(4, "site is not managed by vhost")
		}

		for fileName, fileBytes := range checkPoint.Output {
			currentPath := fmt.Sprintf("%s/%s", vhost.PATH_NGINX_DIR, fileName)
			templatePath := fmt.Sprintf("%s/%s", checkPoint.GetFileName(), fileName)
			currentBytes, errContents := ioutil.ReadFile(currentPath)
			if errContents != nil {
				currentBytes = []byte{}
			}
			edits := myers.ComputeEdits(span.URIFromPath(fileName), string(fileBytes), string(currentBytes))
			diff := fmt.Sprint(gotextdiff.ToUnified(templatePath, currentPath, string(fileBytes), edits))
			fmt.Print(diff)
		}
	},
}

func init() {
	RootCmd.AddCommand(diffCmd)
}
