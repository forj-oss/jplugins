package main

import (
	"github.com/alecthomas/kingpin"
)

type jPluginsApp struct {
	app *kingpin.Application

	listInstalled struct {
		cmd *kingpin.CmdClause
		jenkinsHomePath *string
	}

	checkVersions struct {
		cmd  *kingpin.CmdClause
		jenkinsHomePath *string
	}

	installedPlugins plugins
}

func (a *jPluginsApp) init() {
	a.app = kingpin.New("jplugins", "Jenkins plugins as Code management tool.")

	a.setVersion()

	a.listInstalled.cmd = a.app.Command("list-installed", "Display Jenkins plugins list of current Jenkins installation.")
	a.listInstalled.jenkinsHomePath = a.listInstalled.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default("/var/jenkins_home").String()

	a.checkVersions.cmd = a.app.Command("check-updates", "Display Jenkins plugins which has updates available from existing Jenkins installation.")
	a.checkVersions.jenkinsHomePath = a.checkVersions.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default("/var/jenkins_home").String()
}
