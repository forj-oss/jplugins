package main

import (
	"fmt"
	"log"

	"github.com/forj-oss/forjj-modules/trace"

	"github.com/alecthomas/kingpin"
)

type cmdCheckVersions struct {
	cmd             *kingpin.CmdClause
	jenkinsHomePath *string

	usePreInstalled  *bool
	preInstalledPath *string

	pluginsLock   *string
	usePluginLock *bool

	pluginsFeaturePath *string
	pluginsFeatureFile *string
	usePluginFeature   *bool

	export         *bool
	exportTemplate *string
	exportPath     *string
}

const (
	defaultExportFile = "updates.json"
)

func (c *cmdCheckVersions) init() {
	c.cmd = App.app.Command("check-updates", "Display Jenkins plugins which has updates available from existing Jenkins installation.")
	c.jenkinsHomePath = c.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()

	c.usePreInstalled = c.cmd.Flag("use-pre-installed", "To use pre-installed list file exclusively.").Bool()
	c.preInstalledPath = c.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(defaultJenkinsHome).String()

	c.pluginsLock = c.cmd.Flag("lock-file", "Path to the jplugins.lock file.").Default(".").String()
	c.usePluginLock = c.cmd.Flag("use-lock-file", "To use lock file exclusively.").Bool()

	c.pluginsFeaturePath = c.cmd.Flag("features-path", "Path to the features.lst file.").Default(".").String()
	c.pluginsFeatureFile = c.cmd.Flag("features-filename", "Feature file name.").Default(featureFileName).String()
	c.usePluginFeature = c.cmd.Flag("use-features", "To use features file exclusively.").Bool()

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

	if !c.selectSource() {
		return
	}

	updates := repo.compare(App.installedElements)
	if !*c.export {
		updates.displayUpdates()
	} else {
		export := newPluginsExport(*c.exportPath, *c.exportTemplate, len(updates.plugins))
		if err := export.doItOn(updates); err != nil {
			log.Fatalf("Unable to export. %s", err)
		}
	}
	fmt.Println(len(repo.Plugins), "plugins/groovies loaded.")
	fmt.Println(len(App.installedElements), "plugins/groovies installed.")

}

func (c *cmdCheckVersions) selectSource() (_ bool) {

	if !*c.usePluginLock && !*c.usePreInstalled && !*c.usePluginFeature {
		// load from jenkins
		if App.checkJenkinsHome(*c.jenkinsHomePath) {
			gotrace.Info("Using detected Jenkins home path '%s'", *c.jenkinsHomePath)
			return App.readFromJenkins(*c.jenkinsHomePath)
		}
		if App.checkSimpleFormatFile(*c.pluginsLock, lockFileName) {
			gotrace.Info("Using detected lockfile '%s/%s'", *c.pluginsLock, lockFileName)
			return App.readFromSimpleFormat(*c.pluginsLock, lockFileName)
		}
		if App.checkSimpleFormatFile(*c.preInstalledPath, preInstalledFileName) {
			gotrace.Info("Using detected pre-installed file '%s/%s'", *c.preInstalledPath, preInstalledFileName)
			return App.readFromSimpleFormat(*c.preInstalledPath, preInstalledFileName)
		}
		if App.checkSimpleFormatFile(*c.pluginsFeaturePath, *c.pluginsFeatureFile) {
			gotrace.Info("Using detected feature file '%s/%s'", *c.pluginsFeaturePath, *c.pluginsFeatureFile)
			return App.readFromSimpleFormat(*c.pluginsFeaturePath, *c.pluginsFeatureFile)
		}
		return
	}
	if *c.usePluginLock {
		gotrace.Info("Forcelly using lockfile '%s/%s'", *c.pluginsLock, lockFileName)
		if !App.readFromSimpleFormat(*c.pluginsLock, lockFileName) {
			return
		}

	} else if *c.usePreInstalled {
		gotrace.Info("Forcelly using pre-installed file '%s/%s'", *c.preInstalledPath, preInstalledFileName)
		if !App.readFromSimpleFormat(*c.preInstalledPath, preInstalledFileName) {
			return
		}

	} else {
		gotrace.Info("Forcelly using feature file '%s/%s'", *c.pluginsFeaturePath, *c.pluginsFeatureFile)
		if !App.readFromSimpleFormat(*c.pluginsFeaturePath, *c.pluginsFeatureFile) {
			return
		}
	}
	return true
}
