package main

import (
	"fmt"
	core "jplugins/coremgt"
	"jplugins/utils"
	"log"
	"path"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
)

type cmdCheckVersions struct {
	cmd             *kingpin.CmdClause
	jenkinsHomePath *string
	useJenkinsHome  *bool

	usePreInstalled  *bool
	preInstalledPath *string

	pluginsLock   *string
	usePluginLock *bool

	featureRepoPath    *string
	featureRepoURL     *string
	pluginsFeaturePath *string
	pluginsFeatureFile *string
	usePluginFeature   *bool

	export         *bool
	exportTemplate *string
	exportPath     *string

	updates *core.PluginsStatus
	forcely bool
}

const (
	defaultExportFile = "updates.json"
	jenkinsHomeCheck  = "Jenkins Home"
	lockCheck         = "jplugins lock file"
	featuresCheck     = "jplugins features file"
	preInstallCheck   = "jplugins pre-installed file"
)

func (c *cmdCheckVersions) init() {
	c.cmd = App.app.Command("check-updates", "Display Jenkins plugins which has updates available from existing Jenkins installation.")
	c.jenkinsHomePath = c.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()
	c.useJenkinsHome = c.cmd.Flag("use-jenkins-home", "To use jenkins home plugins list exclusively.").Bool()

	c.usePreInstalled = c.cmd.Flag("use-pre-installed", "To use pre-installed list file exclusively.").Bool()
	c.preInstalledPath = c.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(defaultJenkinsHome).String()

	c.pluginsLock = c.cmd.Flag("lock-file", "Path to the jplugins.lock file.").Default(".").String()
	c.usePluginLock = c.cmd.Flag("use-lock-file", "To use lock file exclusively.").Bool()

	c.featureRepoPath = c.cmd.Flag("features-repo-path", "Path to a feature repository. "+
		"By default, jplugins store the repo clone in jplugins cache directory.").Default(defaultFeaturesRepoPath).String()
	c.featureRepoURL = c.cmd.Flag("features-repo-url", "URL to the feature repository. NOT IMPLEMENTED").Default(defaultFeaturesRepoURL).String()

	c.pluginsFeaturePath = c.cmd.Flag("features-path", "Path to the features.lst file.").String()
	c.pluginsFeatureFile = c.cmd.Flag("features-filename", "Feature file name.").Default(featureFileName).String()
	c.usePluginFeature = c.cmd.Flag("use-features", "To use features file exclusively.").Bool()

	c.export = c.cmd.Flag("export-result", "Export update status to a file.").Bool()
	c.exportPath = c.cmd.Flag("export-as-file", "Full path to the export file to create.").Default(defaultExportFile).String()
	c.exportTemplate = c.cmd.Flag("export-template", "To generate through another custom format.").String()
}

func (c *cmdCheckVersions) doCheckInstalled() {
	choices := c.identifySource()

	App.setJenkinsHome(*c.jenkinsHomePath)

	if err := choices.Run(); err != nil {
		log.Fatalf("Check update issue. %s.", err)
		return
	}
	repo := App.repository

	if !*c.export {
		c.updates.DisplayUpdates()
	} else {
		export := core.NewPluginsExport(*c.exportPath, *c.exportTemplate, c.updates.PluginsLength())
		if err := export.DoItOn(c.updates); err != nil {
			log.Fatalf("Unable to export. %s", err)
		}
	}
	fmt.Println(len(repo.Plugins), "plugins/groovies loaded.")
	fmt.Println(App.installedElements.Length(), "plugins/groovies installed.")
}

// identifySource identify update execution context from file/path existence and forced flags
//
// If no forced flag are given, the following task will be selected if following file/path as described by SetChoice
func (c *cmdCheckVersions) identifySource() (choices *utils.UpdatesSelect) {
	choices = utils.NewUpdatesSelect()

	c.forcely = *c.useJenkinsHome || *c.usePluginLock || *c.usePreInstalled || *c.usePluginFeature

	choices.SetCheck(jenkinsHomeCheck, func() bool {
		state := App.checkJenkinsHome()
		return (!c.forcely || *c.useJenkinsHome) && state
	})

	choices.SetCheck(lockCheck, func() bool {
		state := App.checkSimpleFormatFile(*c.pluginsLock, lockFileName)
		return (!c.forcely || *c.usePluginLock) && state
	})

	choices.SetCheck(preInstallCheck, func() bool {
		state := App.checkSimpleFormatFile(*c.preInstalledPath, preInstalledFileName)
		return (!c.forcely || *c.usePreInstalled) && state
	})

	choices.SetCheck(featuresCheck, func() bool {
		state := App.checkSimpleFormatFile(*c.pluginsFeaturePath, *c.pluginsFeatureFile)
		return (!c.forcely || *c.usePluginFeature) && state
	})

	choices.SetChoice("Checking features, pre-installed and lock files against Jenkins home",
		c.localJenkinsHomeUpdates, jenkinsHomeCheck, lockCheck, featuresCheck, preInstallCheck)

	choices.SetChoice("Checking features and lock files against Jenkins home",
		c.localJenkinsHomeUpdates, jenkinsHomeCheck, lockCheck, featuresCheck)

	choices.SetChoice("Checking lock file against Jenkins home",
		c.localJenkinsHomeUpdates, jenkinsHomeCheck, lockCheck)

	choices.SetChoice("Checking features file against Jenkins home",
		c.localJenkinsHomeUpdates, jenkinsHomeCheck, featuresCheck)

	choices.SetChoice("Checking Jenkins home against Jenkins updates",
		c.jenkinsHomeUpdates, jenkinsHomeCheck)

	choices.SetChoice("Checking pre-installed file against jenkins updates",
		c.jenkinsUpdates, preInstallCheck)

	choices.SetChoice("Checking lock file against jenkins updates",
		c.jenkinsUpdates, lockCheck)

	choices.SetChoice("Checking features file against jenkins updates",
		c.jenkinsUpdates, featuresCheck)

	return
}

// localJenkinsHomeUpdates check local files against Jenkins home
func (c *cmdCheckVersions) localJenkinsHomeUpdates(choice utils.UpdatesSelectChoice, states map[string]bool) error {
	gotrace.Info(choice.Choice)
	return nil
}

// jenkinsHomeUpdates show update of Jenkins Home from Jenkins updates
func (c *cmdCheckVersions) jenkinsHomeUpdates(choice utils.UpdatesSelectChoice, states map[string]bool) error {
	gotrace.Info(choice.Choice)
	App.repository = core.NewRepository()
	repo := App.repository
	if !repo.LoadFromURL() {
		return fmt.Errorf("Issue to load remote repository list")
	}

	elements, err := App.readFromJenkins()
	if err != nil {
		return err
	}

	c.updates = repo.Compare(elements)

	return nil
}

func (c *cmdCheckVersions) jenkinsUpdates(choice utils.UpdatesSelectChoice, states map[string]bool) error {
	gotrace.Info(choice.Choice)

	App.repository = core.NewRepository()
	repo := App.repository
	if !repo.LoadFromURL() {
		return fmt.Errorf("Issue to load remote repository list")
	}

	if states[lockCheck] {
		elements, _ := App.readFromSimpleFormat(*c.pluginsLock, lockFileName)
		c.updates = repo.Compare(elements)
	} else if states[preInstallCheck] {
		elements, _ := App.readFromSimpleFormat(*c.preInstalledPath, preInstalledFileName)
		c.updates = repo.Compare(elements)
	} else if states[featuresCheck] {
		// Load defined features to get plugins list and create a lock data in mem.
		elements, err := App.readFeaturesFromSimpleFormat(*c.featureRepoPath, path.Join(*c.pluginsFeaturePath, *c.pluginsFeatureFile), *c.featureRepoURL)
		if err != nil {
			return fmt.Errorf("Unable to check updates. %s", err)
		}
		c.updates = repo.Compare(elements)
	}

	return nil
}

/*func (c *cmdCheckVersions) selectSource() (_ bool) {
	choices := c.identifySource()

	App.repository = core.NewRepository()
	repo := App.repository
	if !repo.LoadFromURL() {
		return
	}

	App.setJenkinsHome(*c.jenkinsHomePath)
	if !*c.usePluginLock && !*c.usePreInstalled && !*c.usePluginFeature {
		// load from jenkins
		if App.checkJenkinsHome() {
			gotrace.Info("Using detected Jenkins home path '%s'", *c.jenkinsHomePath)
			return App.readFromJenkins()
		}
		// Load from lock file
		if App.checkSimpleFormatFile(*c.pluginsLock, lockFileName) {
			gotrace.Info("Using detected lockfile '%s/%s'", *c.pluginsLock, lockFileName)
			return App.readFromSimpleFormat(*c.pluginsLock, lockFileName)
		}
		// Load from pre-installed
		if App.checkSimpleFormatFile(*c.preInstalledPath, preInstalledFileName) {
			gotrace.Info("Using detected pre-installed file '%s/%s'", *c.preInstalledPath, preInstalledFileName)
			return App.readFromSimpleFormat(*c.preInstalledPath, preInstalledFileName)
		}

		if App.checkSimpleFormatFile(*c.pluginsFeaturePath, *c.pluginsFeatureFile) {
			gotrace.Info("Using detected feature file '%s/%s'", *c.pluginsFeaturePath, *c.pluginsFeatureFile)
			App.readFromSimpleFormat(*c.pluginsFeaturePath, *c.pluginsFeatureFile)
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
		// Load lock file as ref
		if App.checkSimpleFormatFile(*c.pluginsLock, lockFileName) {
			gotrace.Info("Using detected lockfile '%s/%s'", *c.pluginsLock, lockFileName)
			if App.readFromSimpleFormat(*c.pluginsLock, lockFileName) {
				lockData := core.NewPluginsStatus(App.installedElements, repo)

				lockData.ImportInstalled(App.installedElements)

				if !App.readFeatures(*c.featureRepoPath, path.Join(*c.pluginsFeaturePath, *c.pluginsFeatureFile), *c.featureRepoURL, lockData) {
					os.Exit(1)
				}
			}
		}

	}

	return true
}
*/
