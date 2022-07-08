package internal

import (
	"fmt"
	"os"

	"github.com/jetrails/proposal-nginx/sdk/vhost"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:     "create TEMPLATE DOMAIN KEY=VALUE...",
	Short:   "create a vhost from a template",
	Args:    cobra.MinimumNArgs(2),
	PreRunE: ValidateKeyValueArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		domainName := args[1]
		parameters := ParseKeyValueArgs(args[2:])
		if err := vhost.Create(templateName, domainName, parameters); err != nil {
			fmt.Printf("\nError: %s\n\n", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
}
