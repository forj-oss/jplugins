package main

import (
	"fmt"
)

const (
	JenkinsRepo        = "https://updates.jenkins.io"
	JenkinsRepoVersion = "current"
	JenkinsRepoFile    = "update-center.actual.json"
)

func (a *jPluginsApp) doCheckInstalled() {
	a.repository = NewRepository()
	repo := a.repository
	if !repo.loadFrom(JenkinsRepo, JenkinsRepoVersion, JenkinsRepoFile) {
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
