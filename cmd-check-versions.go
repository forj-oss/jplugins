package main

import (
	"fmt"
	"path"

	"github.com/alecthomas/kingpin"
)

type cmdCheckVersions struct {
	cmd              *kingpin.CmdClause
	jenkinsHomePath  *string
	usePreInstalled  *bool
	preInstalledPath *string
	pluginsLock      *string
	usePluginLock    *bool
	export           *bool
	exportTemplate   *string
	exportPath       *string
}

const (
	defaultExportFile = "updates.json"
)

func (c *cmdCheckVersions) init() {
	c.cmd = App.app.Command("check-updates", "Display Jenkins plugins which has updates available from existing Jenkins installation.")
	c.jenkinsHomePath = c.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()
	c.usePreInstalled = c.cmd.Flag("use-pre-installed", "To use pre-installed list instead of jenkins plugins directory.").Bool()
	c.preInstalledPath = c.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(defaultJenkinsHome).String()
	c.pluginsLock = c.cmd.Flag("lock-file", "Path to the jplugins.lock file.").Default(".").String()
	c.usePluginLock = c.cmd.Flag("use-lock-file", "To use lock file instead of jenkins plugins directory.").Bool()
	c.export = c.cmd.Flag("export-result", "Export update status to a file.").Bool()
	c.exportPath = c.cmd.Flag("export-as-file", "Full path to the export file to create.").Default(defaultExportFile).String()
	c.exportTemplate = c.cmd.Flag("export-template", "To generate through another custom format.").String()
}

func (c *cmdCheckVersions) doCheckInstalled() {
	App.repository = NewRepository()
	repo := App.repository
	if !repo.loadFrom() {
		return
	}

	if *c.usePluginLock {
		if !App.readFromSimpleFormat(path.Join(*c.pluginsLock, lockFileName)) {
			return
		}

	} else if *c.usePreInstalled {
		if !App.readFromSimpleFormat(path.Join(*c.preInstalledPath, preInstalledFileName)) {
			return
		}
	} else {
		if !App.readFromJenkins(*c.jenkinsHomePath) {
			return
		}
	}

	updates := repo.compare(App.installedElements)
	if !*c.export {
		updates.displayUpdates()
	} else {
		export := newPluginsExport(*c.exportPath, *c.exportTemplate, len(updates.plugins))
		export.doItOn(updates)
	}
	fmt.Println(len(repo.Plugins), "plugins/groovies loaded.")
	fmt.Println(len(App.installedElements), "plugins/groovies installed.")

}
