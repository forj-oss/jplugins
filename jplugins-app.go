package main

import (
	"github.com/alecthomas/kingpin"
)

type jPluginsApp struct {
	app *kingpin.Application

	listInstalled struct {
		cmd             *kingpin.CmdClause
		jenkinsHomePath *string
		preInstalled    *bool
	}

	checkVersions struct {
		cmd              *kingpin.CmdClause
		jenkinsHomePath  *string
		usePreInstalled  *bool
		preInstalledPath *string
	}

	initCmd struct {
		cmd              *kingpin.CmdClause
		preInstalledPath *string
	}

	installedPlugins plugins
	repository       *repository
}

func (a *jPluginsApp) init() {
	a.app = kingpin.New("jplugins", "Jenkins plugins as Code management tool.")

	a.setVersion()

	a.listInstalled.cmd = a.app.Command("list-installed", "Display Jenkins plugins list of current Jenkins installation.")
	a.listInstalled.jenkinsHomePath = a.listInstalled.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default("/var/jenkins_home").String()
	a.listInstalled.preInstalled = a.listInstalled.cmd.Flag("save-pre-installed", "To create the jplugins-preinstalled.lst instead displaying.").Bool()

	a.checkVersions.cmd = a.app.Command("check-updates", "Display Jenkins plugins which has updates available from existing Jenkins installation.")
	a.checkVersions.jenkinsHomePath = a.checkVersions.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default("/var/jenkins_home").String()
	a.checkVersions.usePreInstalled = a.checkVersions.cmd.Flag("use-pre-installed", "To use pre-installed list instead of jenkins plugins directory.").Bool()
	a.checkVersions.preInstalledPath = a.checkVersions.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default("/var/jenkins_home").String()

	a.initCmd.cmd = a.app.Command("init", "Initialize the 'jplugins.lock' from pre-installed plugins (jplugins-preinstalled.lst).")
	a.initCmd.preInstalledPath = a.initCmd.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(".").String()
}
