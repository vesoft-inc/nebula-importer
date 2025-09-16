package util

import (
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command) error {
	return cmd.Execute()
}
