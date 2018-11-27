package main

import (
	core "jplugins/coremgt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"

	"jplugins/utils"
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
	App.repository = core.NewRepository()
	repo := App.repository
	if !repo.LoadFromURL() {
		os.Exit(1)
	}

	var elements *core.ElementsType

	if utils.CheckFile(*c.preInstalledPath, preInstalledFileName) {
		if e, err := App.readFromSimpleFormat(*c.preInstalledPath, preInstalledFileName); err != nil {
			gotrace.Error("%s", err)
			os.Exit(1)
		} else {
			elements = e
		}
	}

	lockData := core.NewPluginsStatus(elements, repo)

	lockData.ImportInstalled(elements)

	if !App.readFeatures(*c.featureRepoPath, *c.sourceFile, *c.featureRepoURL, lockData) {
		os.Exit(1)
	}

	lockData.DisplayUpdates()

	if !App.writeLockFile(*c.lockFile, lockData) {
		os.Exit(1)
	}

}
