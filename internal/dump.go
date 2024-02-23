package internal

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/null93/vhost/pkg/vhost"
	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump SITE_NAME REVISION",
	Short: "dump to a previous checkpoint",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		siteName := args[0]
		revision := args[1]
		directory, _ := cmd.Flags().GetString("directory")

		revisionInt, errRevision := strconv.Atoi(revision)
		if errRevision != nil {
			ExitWithError(1, fmt.Sprintf("revision %q is not an integer", revision))
		}
		if !vhost.SiteExists(siteName) {
			ExitWithError(2, fmt.Sprintf("site %q does not exist", siteName))
		}

		checkPoint, errCheckPoint := vhost.GetCheckPoint(siteName, revisionInt)
		if errCheckPoint != nil {
			ExitWithError(3, errCheckPoint.Error())
		}

		buf := new(bytes.Buffer)
		gzipWriter := gzip.NewWriter(buf)
		tarWriter := tar.NewWriter(gzipWriter)
		defer tarWriter.Close()
		defer gzipWriter.Close()

		tarFiles, err := checkPoint.GetTarFiles()
		if err != nil {
			ExitWithError(4, err.Error())
		}

		for _, tarFile := range tarFiles {
			if err := tarWriter.WriteHeader(tarFile.Header); err != nil {
				tarWriter.Close()
				gzipWriter.Close()
				ExitWithError(5, err.Error())
			}
			if _, err := tarWriter.Write(tarFile.Body); err != nil {
				tarWriter.Close()
				gzipWriter.Close()
				ExitWithError(6, err.Error())
			}
		}

		if err := tarWriter.Close(); err != nil {
			gzipWriter.Close()
			ExitWithError(7, err.Error())
		}
		if err := gzipWriter.Close(); err != nil {
			ExitWithError(8, err.Error())
		}

		fileName := fmt.Sprintf("%s-%d.tar.gz", siteName, revisionInt)
		filePath := path.Join(directory, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			ExitWithError(9, err.Error())
		}
		defer file.Close()

		_, err = buf.WriteTo(file)
		if err != nil {
			ExitWithError(10, err.Error())
		}

		fmt.Println(filePath)
	},
}

func init() {
	RootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringP("directory", "d", ".", "directory to dump to")
}
