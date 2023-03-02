package util

import (
	"math/rand"
	"time"

	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command) error {
	rand.Seed(time.Now().UnixNano())
	return cmd.Execute()
}
