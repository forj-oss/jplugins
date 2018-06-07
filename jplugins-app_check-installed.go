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
	repo := NewRepository()
	if !repo.loadFrom(JenkinsRepo, JenkinsRepoVersion, JenkinsRepoFile) {
		return
	}

	if !a.readFromJenkins(*a.checkVersions.jenkinsHomePath) {
		return
	}

	updates := repo.compare(a.installedPlugins)
	updates.displayUpdates()
	fmt.Println(len(repo.Plugins), "plugins loaded.")
	fmt.Println(len(a.installedPlugins), "plugins installed.")

}
