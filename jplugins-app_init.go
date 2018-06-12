package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
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

	if !a.readFromSimpleFormat(path.Join(*a.initCmd.preInstalledPath, preInstalledFileName)) {
		return
	}

	lockData := newPluginsStatus(a.installedElements, repo)

	lockData.importInstalled(a.installedElements)

	if !a.readFeatures(*a.initCmd.featureRepoPath, *a.initCmd.sourceFile, *a.initCmd.featureRepoURL, lockData) {
		return
	}

	if !a.writeLockFile(*a.initCmd.lockFile, lockData) {
		return
	}

}

func (a *jPluginsApp) writeLockFile(lockFile string, lockData *pluginsStatus) (_ bool) {

	pluginsList := make([]string, len(lockData.plugins))

	iCount := 0
	for name := range lockData.plugins {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	fd, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		gotrace.Error("Unable to write '%s'. %s", lockFileName, err)
		return
	}
	defer fd.Close()

	for _, name := range pluginsList {
		plugin := lockData.plugins[name]
		fmt.Fprintf(fd, "plugin:%s:%s\n", name, plugin.newVersion)
	}

	pluginsList = make([]string, len(lockData.groovies))

	iCount = 0
	for name := range lockData.groovies {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	for _, name := range pluginsList {
		groovy := lockData.groovies[name]
		fmt.Fprintf(fd, "groovy:%s:%s\n", name, groovy.newCommit)
	}

	gotrace.Info("%s written\n", lockFileName)
	return true
}

func (a *jPluginsApp) readFeatures(featurePath, featureFile, featureURL string, lockData *pluginsStatus) (_ bool) {
	gotrace.Trace("Loading constraints...")
	if gotrace.IsInfoMode() {
		fmt.Printf("Reading %s\n--------\n", featureFileName)
	}
	fd, err := os.Open(featureFile)
	if err != nil {
		gotrace.Error("Unable to read '%s'. %s", featureFileName, err)
		return
	}

	if featurePath != defaultFeaturesRepoPath {
		lockData.setLocal()
	}
	lockData.setFeaturesPath(featurePath)
	lockData.setFeaturesRepoURL(featureURL)

	fileScan := bufio.NewScanner(fd)
	for fileScan.Scan() {
		line := strings.Trim(fileScan.Text(), " \n")
		if gotrace.IsInfoMode() {
			fmt.Printf("== %s ==\n", line)
		}
		lockData.checkElement(line, func(ftype, name, version string) {
			switch ftype {
			case "feature":
				lockData.checkFeature(name)
			//case "groovy":
			case "plugin":
				lockData.checkPlugin(name, version, nil)
			default:
				gotrace.Warning("feature type '%s' is currently not supported. Ignored.", ftype)
				return
			}
		})
	}

	if gotrace.IsInfoMode() {
		fmt.Println("--------")
	}
	gotrace.Trace("Identifying version from constraints...")
	lockData.definePluginsVersion()

	if !lockData.checkMinDep() {
		return
	}

	lockData.displayUpdates()

	return true
}
