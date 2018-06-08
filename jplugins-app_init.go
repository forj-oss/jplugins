package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
)

const (
	lockFileName    = "jplugins.lock"
	featureFileName = "jplugins.lst"
)

// doInit read `jplugins-preinstalled.lst` and `jplugins-features.yaml` to create a lock file
func (a *jPluginsApp) doInit() {
	a.repository = NewRepository()
	repo := a.repository
	if !repo.loadFrom() {
		return
	}

	if !a.readFromPreInstalled(*a.initCmd.preInstalledPath) {
		return
	}

	lockData := newPluginsStatus(a.installedPlugins, repo)

	lockData.importInstalled(a.installedPlugins)

	if !a.readFeatures(lockData) {
		return
	}

	if !a.writeLockFile(lockData) {
		return
	}

}

func (a *jPluginsApp) writeLockFile(lockData *pluginsStatus) (_ bool) {

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

func (a *jPluginsApp) readFeatures(lockData *pluginsStatus) (_ bool) {
	gotrace.Trace("Loading constraints...")
	if gotrace.IsInfoMode() {
		fmt.Printf("Reading %s\n--------\n", featureFileName)
	}
	fd, err := os.Open(featureFileName)
	if err != nil {
		gotrace.Error("Unable to read '%s'. %s", featureFileName, err)
		return
	}

	fileScan := bufio.NewScanner(fd)
	for fileScan.Scan() {
		line := strings.Trim(fileScan.Text(), " \n")
		if line[0] == '#' {
			continue
		}
		if gotrace.IsInfoMode() {
			fmt.Printf("== %s ==\n", line)
		}
		lockData.checkElement(line)
	}

	if gotrace.IsInfoMode() {
		fmt.Println("--------")
	}
	gotrace.Trace("Identifying version from constraints...")
	lockData.definePluginsVersion()

	lockData.displayUpdates()

	return true
}
