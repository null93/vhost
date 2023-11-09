package internal

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/jetrails/proposal-nginx/pkg/vhost"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:     "upgrade SITE_NAME KEY=VALUE...",
	Short:   "upgrade a vhost from a template",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: ValidateKeyValueArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		siteName := args[0]
		templateInput := ParseKeyValueArgs(args[1:])

		if !vhost.SiteExists(siteName) {
			ExitWithError(1, "site does not exist")
		}

		checkPoint, errCheckPoint := vhost.GetLatestCheckPoint(siteName)
		if errCheckPoint != nil {
			ExitWithError(2, errCheckPoint.Error())
		}

		latestTemplate, errTemplate := vhost.LoadTemplate(checkPoint.Template.Name)
		if errTemplate != nil {
			ExitWithError(3, errTemplate.Error())
		}

		mergedInput := MergeInput(checkPoint.Input, templateInput)

		newCheckPoint, errNewCheckPoint := vhost.NewCheckPoint(siteName, latestTemplate, mergedInput)
		if errNewCheckPoint != nil {
			ExitWithError(4, errNewCheckPoint.Error())
		}

		if dryRun {
			fullDiff := ""
			seen := map[string]bool{}
			for fileName, fileBytes := range checkPoint.Output {
				seen[fileName] = true
				currentPath := fmt.Sprintf("%s/%s", vhost.PATH_NGINX_DIR, fileName)
				templatePath := fmt.Sprintf("%s/%s", vhost.PATH_NGINX_DIR, fileName)
				currentBytes, errContents := ioutil.ReadFile(currentPath)
				if errContents != nil {
					currentBytes = []byte{}
				}
				edits := myers.ComputeEdits(span.URIFromPath(fileName), string(currentBytes), string(fileBytes))
				diff := fmt.Sprint(gotextdiff.ToUnified("old"+templatePath, "new"+currentPath, string(currentBytes), edits))
				fullDiff = fullDiff + diff
			}
			if len(fullDiff) > 0 {
				ExitWithError(5, "site deviates from template in checkpoint, rollback first")
			}
			for fileName, fileBytes := range newCheckPoint.Output {
				delete(seen, fileName)
				currentPath := fmt.Sprintf("%s/%s", vhost.PATH_NGINX_DIR, fileName)
				templatePath := fmt.Sprintf("%s/%s", vhost.PATH_NGINX_DIR, fileName)
				currentBytes := checkPoint.Output[fileName]
				edits := myers.ComputeEdits(span.URIFromPath(fileName), string(currentBytes), string(fileBytes))
				diff := fmt.Sprint(gotextdiff.ToUnified("old"+templatePath, "new"+currentPath, string(currentBytes), edits))
				fullDiff = fullDiff + diff
			}
			for fileName := range seen {
				currentPath := fmt.Sprintf("%s/%s", vhost.PATH_NGINX_DIR, fileName)
				templatePath := fmt.Sprintf("%s/%s", vhost.PATH_NGINX_DIR, fileName)
				fileBytes := newCheckPoint.Output[fileName]
				currentBytes := checkPoint.Output[fileName]
				edits := myers.ComputeEdits(span.URIFromPath(fileName), string(currentBytes), string(fileBytes))
				diff := fmt.Sprint(gotextdiff.ToUnified("old"+templatePath, "new"+currentPath, string(currentBytes), edits))
				fullDiff = fullDiff + diff
			}
			fmt.Print(fullDiff)
			return
		}

		if checkPoint.Template.Hash() == latestTemplate.Hash() {
			ExitWithError(7, "template is already up to date")
		}

		if errDelete := checkPoint.Output.DeleteFiles(true); errDelete != nil {
			ExitWithError(8, errDelete.Error())
		}

		newCheckPoint.Revision = checkPoint.Revision + 1
		newCheckPoint.Description = "upgraded template"
		newCheckPoint.Timestamp = time.Now()
		newCheckPoint.Template = latestTemplate
		newCheckPoint.Input = mergedInput

		errOutputSave := newCheckPoint.Output.Save()
		if errOutputSave != nil {
			ExitWithError(9, errOutputSave.Error())
		}

		errSave := newCheckPoint.Save()
		if errSave != nil {
			ExitWithError(10, errSave.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolP("dry-run", "d", false, "show diff, don't make changes")
}
