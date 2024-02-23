package internal

import (
	"fmt"

	"github.com/null93/vhost/pkg/utils"
	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history SITE_NAME",
	Short: "history of a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]
		if !vhost.SiteExists(siteName) {
			ExitWithError(1, fmt.Sprintf("site %q does not exist", siteName))
		}
		tbl := utils.NewTable(
			"Revision",
			"Timestamp",
			"Status",
			"Template",
			"Input",
			"Output",
			"Description",
		)
		checkPoints, errCheckPoints := vhost.GetCheckPoints(siteName)
		if errCheckPoints != nil {
			ExitWithError(2, errCheckPoints.Error())
		}
		lastHashes := [][]string{}
		for revision, checkPoint := range checkPoints {
			state := "superseded"
			if revision == len(checkPoints)-1 {
				state = "current"
			}
			templateHash := checkPoint.Template.Hash()
			inputHash := checkPoint.Input.Hash()
			outputHash := checkPoint.Output.Hash()
			if revision > 0 && lastHashes[revision-1][0] == templateHash {
				templateHash = "^^^^^^^"
			}
			if revision > 0 && lastHashes[revision-1][1] == inputHash {
				inputHash = "^^^^^^^"
			}
			if revision > 0 && lastHashes[revision-1][2] == outputHash {
				outputHash = "^^^^^^^"
			}
			tbl.AddRow(
				fmt.Sprintf("%d", revision+1),
				checkPoint.Timestamp.Format("01/02/2006 15:04:05"),
				state,
				templateHash,
				inputHash,
				outputHash,
				checkPoint.Description,
			)
			lastHashes = append(lastHashes, []string{
				checkPoint.Template.Hash(),
				checkPoint.Input.Hash(),
				checkPoint.Output.Hash(),
			})
		}
		tbl.PrintSeparator()
		tbl.Print()
		tbl.PrintSeparator()
	},
}

func init() {
	RootCmd.AddCommand(historyCmd)
}
