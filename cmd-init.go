package main

import (
	"path"

	"github.com/alecthomas/kingpin"
)

type cmdInit struct {
	cmd              *kingpin.CmdClause
	preInstalledPath *string
	sourceFile       *string
	lockFile         *string
	featureRepoPath  *string
	featureRepoURL   *string
}

// doInit read `jplugins-preinstalled.lst` and `jplugins-features.yaml` to create a lock file
func (c *cmdInit) doInit() {
	App.repository = NewRepository()
	repo := App.repository
	if !repo.loadFrom() {
		return
	}

	if !App.readFromSimpleFormat(path.Join(*c.preInstalledPath, preInstalledFileName)) {
		return
	}

	lockData := newPluginsStatus(App.installedElements, repo)

	lockData.importInstalled(App.installedElements)

	if !App.readFeatures(*c.featureRepoPath, *c.sourceFile, *c.featureRepoURL, lockData) {
		return
	}

	if !App.writeLockFile(*c.lockFile, lockData) {
		return
	}

}
