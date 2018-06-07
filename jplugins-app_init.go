package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/forj-oss/forjj-modules/trace"
)

const (
	lockFileName = "jplugins.lock"
)

// doInit read `jplugins-preinstalled.lst` and `jplugins-features.yaml` to create a lock file
func (a *jPluginsApp) doInit() {
	a.repository = NewRepository()
	repo := a.repository
	if !repo.loadFrom(JenkinsRepo, JenkinsRepoVersion, JenkinsRepoFile) {
		return
	}

	if !a.readFromPreInstalled(*a.initCmd.preInstalledPath) {
		return
	}
	if !a.writeLockFile() {
		return
	}

}

func (a *jPluginsApp) writeLockFile() (_ bool) {

	pluginsList := make([]string, len(a.installedPlugins))

	iCount := 0
	for name := range a.installedPlugins {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	fd, err := os.OpenFile(lockFileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		gotrace.Error("Unable to write '%s'. %s", lockFileName, err)
		return
	}
	defer fd.Close()

	for name, plugin := range a.installedPlugins {
		fmt.Fprintf(fd, "plugin:%s:%s\n", name, plugin.Version)
	}

	gotrace.Info("%s written\n", lockFileName)
	return true
}
