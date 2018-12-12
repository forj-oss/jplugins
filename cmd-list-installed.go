package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
)

type cmdListInstalled struct {
	cmd             *kingpin.CmdClause
	jenkinsHomePath *string
	preInstalled    *bool
}

func (c *cmdListInstalled) doListInstalled() {
	App.setJenkinsHome(*c.jenkinsHomePath)
	elements, err := App.readFromJenkins()
	if err != nil {
		gotrace.Error("%s", err)
		os.Exit(1)
	}
	if *App.listInstalled.preInstalled {
		if !App.saveVersionAsPreInstalled(*c.jenkinsHomePath, elements) {
			os.Exit(1)
		}
		return
	}
	App.printOutVersion(elements)
}
