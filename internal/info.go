package internal

import (
	"fmt"
	"os"

	. "github.com/jetrails/jrctl/pkg/output"

	"github.com/jetrails/proposal-nginx/sdk/vhost"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info DOMAIN",
	Short:   "info a vhost given the domain name",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domainName := args[0]

		if status := vhost.Info(domainName); status != nil {
			tbl := NewTable(Columns{" "," "})
			enabledMap := map[bool]string{true: "enabled", false: "disabled"}
			hash := status.VirtualHost.Template.Hash ()
			tbl.AddRow(Columns{"enabled:", enabledMap[status.Enabled]})
			tbl.AddRow(Columns{"template-hash:", hash[:8] + "..." + hash[len(hash)-8:]})
			for key, value := range status.VirtualHost.Input {
				tbl.AddRow(Columns{key + ":", value})
			}
			tbl.PrintTable()
			fmt.Println()
		} else {
			fmt.Printf ("\nError: could not find config with name %q\n\n", domainName)
			os.Exit (1)
		}

	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
}
