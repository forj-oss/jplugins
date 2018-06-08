package main

import (
	"fmt"
)


func (a *jPluginsApp) doCheckInstalled() {
	a.repository = NewRepository()
	repo := a.repository
	if !repo.loadFrom() {
		return
	}

	if *a.checkVersions.usePreInstalled {
		if !a.readFromPreInstalled(*a.checkVersions.preInstalledPath) {
			return
		}
	} else {
		if !a.readFromJenkins(*a.checkVersions.jenkinsHomePath) {
			return
		}
	}

	updates := repo.compare(a.installedPlugins)
	updates.displayUpdates()
	fmt.Println(len(repo.Plugins), "plugins loaded.")
	fmt.Println(len(a.installedPlugins), "plugins installed.")

}
