package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

type cmdInitLockfile struct {
	cmd              *kingpin.CmdClause
	preInstalledPath *string
	sourceFile       *string
	lockFile         *string
	featureRepoPath  *string
	featureRepoURL   *string
}

func (c *cmdInitLockfile) init(parent *kingpin.CmdClause) {
	c.cmd = parent.Command("lockfile", "Initialize the 'jplugins.lock' from pre-installed plugins (jplugins-preinstalled.lst) and features file.").Default()
	c.preInstalledPath = c.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(".").String()
	c.sourceFile = c.cmd.Flag("feature-file", "Full path to a feature file.").Default(featureFileName).String()
	c.lockFile = c.cmd.Flag("lock-file", "Full path to the lock file.").Default(lockFileName).String()
	c.featureRepoPath = c.cmd.Flag("features-repo-path", "Path to a feature repository. "+
		"By default, jplugins store the repo clone in jplugins cache directory.").Default(defaultFeaturesRepoPath).String()
	c.featureRepoURL = c.cmd.Flag("features-repo-url", "URL to the feature repository. NOT IMPLEMENTED").Default(defaultFeaturesRepoURL).String()

}

func (c *cmdInitLockfile) DoInitLockfile() {
	App.repository = NewRepository()
	repo := App.repository
	if !repo.loadFrom() {
		os.Exit(1)
	}

	if !App.readFromSimpleFormat(*c.preInstalledPath, preInstalledFileName) {
		os.Exit(1)
	}

	lockData := newPluginsStatus(App.installedElements, repo)

	lockData.importInstalled(App.installedElements)

	if !App.readFeatures(*c.featureRepoPath, *c.sourceFile, *c.featureRepoURL, lockData) {
		os.Exit(1)
	}

	if !App.writeLockFile(*c.lockFile, lockData) {
		os.Exit(1)
	}

}
