package internal

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func ParseKeyValueArgs(args []string) map[string]string {
	mapping := map[string]string{}
	for _, arg := range args {
		parts := strings.Split(arg, "=")
		mapping[parts[0]] = parts[1]
	}
	return mapping
}

func ValidateKeyValueArgs(after int) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, arg := range args[after:] {
			parts := strings.Split(arg, "=")
			if len(parts) != 2 {
				return fmt.Errorf("invalid key value pair passed as argument %q", arg)
			}
		}
		return nil
	}
}

var RootCmd = &cobra.Command{
	Use:     "vhost",
	Version: "0.0.0",
	Short:   "manage nginx virtual hosts",
}

func init() {
	RootCmd.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true})
	RootCmd.CompletionOptions.DisableDefaultCmd = true
}
