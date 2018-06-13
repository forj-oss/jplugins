package main

import (
	"github.com/alecthomas/kingpin"
)

type cmdListInstalled struct {
	cmd             *kingpin.CmdClause
	jenkinsHomePath *string
	preInstalled    *bool
}

func (c *cmdListInstalled) doListInstalled() {
	if !App.readFromJenkins(*c.jenkinsHomePath) {
		return
	}
	if *App.listInstalled.preInstalled {
		App.saveVersionAsPreInstalled(*c.jenkinsHomePath, App.installedElements)
		return
	}
	App.printOutVersion(App.installedElements)
}
