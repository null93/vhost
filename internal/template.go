package internal

import (
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "explore available templates",
}

func init() {
	RootCmd.AddCommand(templateCmd)
}
