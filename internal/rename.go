package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename SITE_NAME NEW_SITE_NAME",
	Short: "rename a site managed by vhost",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		oldSiteName := args[0]
		newSiteName := args[1]

		if oldSiteName == newSiteName {
			ExitWithError(1, "new site name must differ from old site name")
		}

		if !vhost.SiteExists(oldSiteName) {
			ExitWithError(2, fmt.Sprintf("site %q does not exist", oldSiteName))
		}

		if vhost.SiteExists(newSiteName) {
			ExitWithError(3, fmt.Sprintf("site %q already exists", newSiteName))
		}

		checkPoints, errCheckPoints := vhost.GetCheckPoints(oldSiteName)
		if errCheckPoints != nil {
			ExitWithError(4, errCheckPoints.Error())
		}

		if len(checkPoints) < 1 {
			ExitWithError(5, fmt.Sprintf("site %q has no checkpoints", oldSiteName))
		}

		latestCheckPoint := checkPoints[len(checkPoints)-1]

		if !latestCheckPoint.Template.Exists() {
			ExitWithError(6, "site is not managed by vhost")
		}

		for fileName, fileBytes := range latestCheckPoint.Output {
			currentPath := path.Join(vhost.PATH_NGINX_DIR, fileName)
			currentBytes, errRead := ioutil.ReadFile(currentPath)
			if errRead != nil {
				ExitWithError(7, fmt.Sprintf("missing file %q", currentPath))
			}
			if string(currentBytes) != string(fileBytes) {
				ExitWithError(8, "site deviates from template in checkpoint, rollback first")
			}
		}

		site, errSite := vhost.GetSite(oldSiteName)
		if errSite != nil {
			ExitWithError(9, errSite.Error())
		}
		wasEnabled := site.State == vhost.SITE_ENABLED

		newInput := vhost.TemplateInput{}
		for k, v := range latestCheckPoint.Input {
			newInput[k] = v
		}
		newInput["site_name"] = newSiteName
		newOutput, errRender := latestCheckPoint.Template.Render(newSiteName, newInput)
		if errRender != nil {
			ExitWithError(10, errRender.Error())
		}

		if errDelete := latestCheckPoint.Output.DeleteFiles(true); errDelete != nil {
			ExitWithError(11, errDelete.Error())
		}
		if wasEnabled {
			if errDisable := vhost.DisableSite(oldSiteName); errDisable != nil {
				ExitWithError(12, errDisable.Error())
			}
		}

		for _, cp := range checkPoints {
			oldPath := path.Join(vhost.PATH_CHECKPOINTS_DIR, fmt.Sprintf("%s_%d.state", oldSiteName, cp.Revision))
			newPath := path.Join(vhost.PATH_CHECKPOINTS_DIR, fmt.Sprintf("%s_%d.state", newSiteName, cp.Revision))
			if errRename := os.Rename(oldPath, newPath); errRename != nil {
				ExitWithError(13, errRename.Error())
			}
		}

		renameCheckPoint := vhost.CheckPoint{
			Revision:    latestCheckPoint.Revision + 1,
			SiteName:    newSiteName,
			Timestamp:   time.Now(),
			Template:    latestCheckPoint.Template,
			Input:       newInput,
			Output:      newOutput,
			Description: fmt.Sprintf("renamed from %q to %q", oldSiteName, newSiteName),
		}
		if errSave := renameCheckPoint.Save(); errSave != nil {
			ExitWithError(14, errSave.Error())
		}

		if errOutputSave := renameCheckPoint.Output.Save(); errOutputSave != nil {
			ExitWithError(15, errOutputSave.Error())
		}

		if wasEnabled {
			if errEnable := vhost.EnableSite(newSiteName); errEnable != nil {
				ExitWithError(16, errEnable.Error())
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(renameCmd)
}
