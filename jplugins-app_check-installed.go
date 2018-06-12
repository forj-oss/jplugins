package main

import (
	"path"
	"fmt"
)


func (a *jPluginsApp) doCheckInstalled() {
	a.repository = NewRepository()
	repo := a.repository
	if !repo.loadFrom() {
		return
	}

	// TODO: Read from lock file
	if *a.checkVersions.usePreInstalled {
		if !a.readFromSimpleFormat(path.Join(*a.initCmd.preInstalledPath, preInstalledFileName)) {
			return
		}
	} else {
		if !a.readFromJenkins(*a.checkVersions.jenkinsHomePath) {
			return
		}
	}

	updates := repo.compare(a.installedElements)
	updates.displayUpdates()
	fmt.Println(len(repo.Plugins), "plugins/groovies loaded.")
	fmt.Println(len(a.installedElements), "plugins/groovies installed.")

}
