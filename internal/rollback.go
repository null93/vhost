package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback SITE_NAME REVISION",
	Short: "rollback to a previous checkpoint",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]
		revision := args[1]
		discard, _ := cmd.Flags().GetBool("discard")
		save, _ := cmd.Flags().GetBool("save")

		if discard && save {
			ExitWithError(1, "--discard and --save are mutually exclusive")
		}

		revisionInt, errRevision := strconv.Atoi(revision)
		if errRevision != nil {
			ExitWithError(1, fmt.Sprintf("revision %q is not an integer", revision))
		}
		if !vhost.SiteExists(siteName) {
			ExitWithError(2, fmt.Sprintf("site %q does not exist", siteName))
		}

		latestCheckPoint, errLatest := vhost.GetLatestCheckPoint(siteName)
		if errLatest != nil {
			ExitWithError(3, errLatest.Error())
		}

		if !latestCheckPoint.Template.Exists() {
			ExitWithError(4, "site is not managed by vhost")
		}

		deviationsFound := false
		deviatedFiles := map[string][]byte{}
		for fileName, fileBytes := range latestCheckPoint.Output {
			currentPath := path.Join(vhost.PATH_NGINX_DIR, fileName)
			currentBytes, errRead := ioutil.ReadFile(currentPath)
			if errRead != nil {
				currentBytes = []byte{}
			}
			deviatedFiles[fileName] = currentBytes
			if string(currentBytes) != string(fileBytes) {
				deviationsFound = true
			}
		}
		if deviationsFound && !discard && !save {
			ExitWithError(15, fmt.Sprintf("there are manual changes that deviate from the template, check 'vhost diff %s' for details; pass --discard to drop them or --save to capture them as a checkpoint", siteName))
		}

		if save && deviationsFound {
			frozenTemplate := latestCheckPoint.Template
			frozenTemplate.Files = map[string][]byte{}
			snapshotOutput := vhost.TemplateOutput{}
			for fileName, fileBytes := range deviatedFiles {
				frozenTemplate.Files[fileName] = fileBytes
				snapshotOutput[fileName] = fileBytes
			}
			snapshot := vhost.CheckPoint{
				Revision:    latestCheckPoint.Revision + 1,
				SiteName:    latestCheckPoint.SiteName,
				Timestamp:   time.Now(),
				Template:    frozenTemplate,
				Input:       latestCheckPoint.Input,
				Output:      snapshotOutput,
				Description: "deviations saved",
				ManualEdits: true,
			}
			if errSave := snapshot.Save(); errSave != nil {
				ExitWithError(16, errSave.Error())
			}
			latestCheckPoint = snapshot
		}

		checkPoint, errCheckPoint := vhost.GetCheckPoint(siteName, revisionInt)
		if errCheckPoint != nil {
			ExitWithError(5, errCheckPoint.Error())
		}

		targetSiteName := checkPoint.SiteName
		if targetSiteName == "" {
			targetSiteName = siteName
		}
		crossRename := targetSiteName != siteName

		if crossRename && vhost.SiteExists(targetSiteName) {
			ExitWithError(6, fmt.Sprintf("cannot rollback across rename: site %q already exists", targetSiteName))
		}

		site, errSite := vhost.GetSite(siteName)
		if errSite != nil {
			ExitWithError(7, errSite.Error())
		}
		wasEnabled := site.State == vhost.SITE_ENABLED

		if errDelete := latestCheckPoint.Output.DeleteFiles(true); errDelete != nil {
			ExitWithError(8, errDelete.Error())
		}

		if crossRename {
			if wasEnabled {
				if errDisable := vhost.DisableSite(siteName); errDisable != nil {
					ExitWithError(9, errDisable.Error())
				}
			}

			allCheckpoints, errAll := vhost.GetCheckPoints(siteName)
			if errAll != nil {
				ExitWithError(10, errAll.Error())
			}
			for _, cp := range allCheckpoints {
				oldPath := path.Join(vhost.PATH_CHECKPOINTS_DIR, fmt.Sprintf("%s_%d.state", siteName, cp.Revision))
				newPath := path.Join(vhost.PATH_CHECKPOINTS_DIR, fmt.Sprintf("%s_%d.state", targetSiteName, cp.Revision))
				if errRename := os.Rename(oldPath, newPath); errRename != nil {
					ExitWithError(11, errRename.Error())
				}
			}
		}

		checkPoint.Revision = latestCheckPoint.Revision + 1
		if crossRename {
			checkPoint.Description = fmt.Sprintf("rollback to revision %d (renamed from %q to %q)", revisionInt, siteName, targetSiteName)
		} else {
			checkPoint.Description = fmt.Sprintf("rollback to revision %d", revisionInt)
		}
		if deviationsFound && discard {
			checkPoint.Description += " (deviations abandoned)"
		}
		checkPoint.Timestamp = time.Now()

		if errOutputSave := checkPoint.Output.Save(); errOutputSave != nil {
			ExitWithError(12, errOutputSave.Error())
		}
		if errSave := checkPoint.Save(); errSave != nil {
			ExitWithError(13, errSave.Error())
		}

		if crossRename && wasEnabled {
			if errEnable := vhost.EnableSite(targetSiteName); errEnable != nil {
				ExitWithError(14, errEnable.Error())
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(rollbackCmd)
	rollbackCmd.Flags().BoolP("discard", "d", false, "discard manual changes that deviate from the latest checkpoint")
	rollbackCmd.Flags().BoolP("save", "s", false, "save manual changes as a new checkpoint before rolling back")
}
