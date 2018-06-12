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
		sourceFile       *string
		lockFile         *string
		featureRepoPath  *string
		featureRepoURL   *string
	}

	installCmd struct {
		cmd             *kingpin.CmdClause
		lockFile        *string
		featureRepoPath *string
		featureRepoURL  *string
		jenkinsHomePath *string
	}

	installedElements plugins
	repository        *repository
}

const (
	defaultFeaturesRepoName = "jenkins-install-inits"
	defaultFeaturesRepoPath = ".jplugins/repo-cache" + defaultFeaturesRepoName
	defaultFeaturesRepoURL  = "https://github.com/forj-oss/" + defaultFeaturesRepoName
	defaultJenkinsHome      = "/var/jenkins_home"
)

func (a *jPluginsApp) init() {
	a.app = kingpin.New("jplugins", "Jenkins plugins as Code management tool.")

	a.setVersion()

	a.listInstalled.cmd = a.app.Command("list-installed", "Display Jenkins plugins list of current Jenkins installation.")
	a.listInstalled.jenkinsHomePath = a.listInstalled.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()
	a.listInstalled.preInstalled = a.listInstalled.cmd.Flag("save-pre-installed", "To create the jplugins-preinstalled.lst instead displaying.").Bool()

	a.checkVersions.cmd = a.app.Command("check-updates", "Display Jenkins plugins which has updates available from existing Jenkins installation.")
	a.checkVersions.jenkinsHomePath = a.checkVersions.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()
	a.checkVersions.usePreInstalled = a.checkVersions.cmd.Flag("use-pre-installed", "To use pre-installed list instead of jenkins plugins directory.").Bool()
	a.checkVersions.preInstalledPath = a.checkVersions.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(defaultJenkinsHome).String()

	a.initCmd.cmd = a.app.Command("init", "Initialize the 'jplugins.lock' from pre-installed plugins (jplugins-preinstalled.lst).")
	a.initCmd.preInstalledPath = a.initCmd.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(".").String()
	a.initCmd.sourceFile = a.initCmd.cmd.Flag("feature-file", "Full path to a feature file.").Default(featureFileName).String()
	a.initCmd.lockFile = a.initCmd.cmd.Flag("lock-file", "Full path to the lock file.").Default(lockFileName).String()
	a.initCmd.featureRepoPath = a.initCmd.cmd.Flag("features-repo-path", "Path to a feature repository. "+
		"By default, jplugins store the repo clone in jplugins cache directory.").Default(defaultFeaturesRepoPath).String()
	a.initCmd.featureRepoURL = a.initCmd.cmd.Flag("features-repo-url", "URL to the feature repository. NOT IMPLEMENTED").Default(defaultFeaturesRepoURL).String()

	a.installCmd.cmd = a.app.Command("install", "Install plugins and groovies defined by the 'jplugins.lock'.")
	a.installCmd.lockFile = a.installCmd.cmd.Flag("lock-file", "Full path to the lock file.").Default(lockFileName).String()
	a.installCmd.featureRepoPath = a.installCmd.cmd.Flag("features-repo-path", "Path to a feature repository. "+
		"By default, jplugins store the repo clone in jplugins cache directory.").Default(defaultFeaturesRepoPath).String()
	a.installCmd.featureRepoURL = a.installCmd.cmd.Flag("features-repo-url", "URL to the feature repository. NOT IMPLEMENTED").Default(defaultFeaturesRepoURL).String()
	a.installCmd.jenkinsHomePath = a.checkVersions.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()
}
