package main

import (
	"github.com/alecthomas/kingpin"
)

type jPluginsApp struct {
	app *kingpin.Application

	listInstalled *kingpin.CmdClause
	jenkinsHomePath *string

	installedPlugins plugins
}

func (a *jPluginsApp) init() {
	a.app = kingpin.New("jplugins", "Jenkins plugins as Code management tool.")

	a.setVersion()

	a.listInstalled = a.app.Command("list-installed", "Display Jenkins plugins list of current Jenkins installation.")
	a.jenkinsHomePath = a.listInstalled.Flag("jenkins-home", "Where Jenkins is installed.").Default("/var/jenkins_home").String()
}
