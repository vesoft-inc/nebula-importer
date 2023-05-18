package main

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/cmd/util"
)

func main() {
	command := cmd.NewDefaultImporterCommand()
	if err := util.Run(command); err != nil {
		util.CheckErr(err)
	}
}
