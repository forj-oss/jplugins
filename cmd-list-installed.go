package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

type cmdListInstalled struct {
	cmd             *kingpin.CmdClause
	jenkinsHomePath *string
	preInstalled    *bool
}

func (c *cmdListInstalled) doListInstalled() {
	if !App.readFromJenkins(*c.jenkinsHomePath) {
		os.Exit(1)
	}
	if *App.listInstalled.preInstalled {
		App.saveVersionAsPreInstalled(*c.jenkinsHomePath, App.installedElements)
		os.Exit(1)
	}
	App.printOutVersion(App.installedElements)
}
