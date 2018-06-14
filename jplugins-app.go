package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	git "github.com/forj-oss/go-git"
	"github.com/forj-oss/utils"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
	yaml "gopkg.in/yaml.v2"
)

type jPluginsApp struct {
	app           *kingpin.Application
	listInstalled cmdListInstalled
	checkVersions cmdCheckVersions
	initCmd       cmdInit
	installCmd    cmdInstall

	installedElements plugins
	repository        *repository
}

const (
	defaultFeaturesRepoName = "jenkins-install-inits"
	defaultFeaturesRepoPath = ".jplugins/repo-cache/" + defaultFeaturesRepoName
	defaultFeaturesRepoURL  = "https://github.com/forj-oss/" + defaultFeaturesRepoName
	defaultJenkinsHome      = "/var/jenkins_home"
	lockFileName            = "jplugins.lock"
	featureFileName         = "jplugins.lst"
	preInstalledFileName    = "jplugins-preinstalled.lst"
)

var build_branch, build_commit, build_date, build_tag string

func (a *jPluginsApp) init() {
	a.app = kingpin.New("jplugins", "Jenkins plugins as Code management tool.")

	a.setVersion()

	a.listInstalled.cmd = a.app.Command("list-installed", "Display Jenkins plugins list of current Jenkins installation.")
	a.listInstalled.jenkinsHomePath = a.listInstalled.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()
	a.listInstalled.preInstalled = a.listInstalled.cmd.Flag("save-pre-installed", "To create the jplugins-preinstalled.lst instead displaying.").Bool()

	a.checkVersions.cmd = a.app.Command("check-updates", "Display Jenkins plugins which has updates available from existing Jenkins installation.")
	a.checkVersions.jenkinsHomePath = a.checkVersions.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()
	a.checkVersions.usePreInstalled = a.checkVersions.cmd.Flag("use-pre-installed", "To use pre-installed list instead of jenkins plugins directory.").Bool()
	a.checkVersions.preInstalledPath = a.checkVersions.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(defaultJenkinsHome).String()

	a.initCmd.cmd = a.app.Command("init", "Initialize the 'jplugins.lock' from pre-installed plugins (jplugins-preinstalled.lst).")
	a.initCmd.preInstalledPath = a.initCmd.cmd.Flag("pre-installed-path", "Path to the pre-installed.lst file.").Default(".").String()
	a.initCmd.sourceFile = a.initCmd.cmd.Flag("feature-file", "Full path to a feature file.").Default(featureFileName).String()
	a.initCmd.lockFile = a.initCmd.cmd.Flag("lock-file", "Full path to the lock file.").Default(lockFileName).String()
	a.initCmd.featureRepoPath = a.initCmd.cmd.Flag("features-repo-path", "Path to a feature repository. "+
		"By default, jplugins store the repo clone in jplugins cache directory.").Default(defaultFeaturesRepoPath).String()
	a.initCmd.featureRepoURL = a.initCmd.cmd.Flag("features-repo-url", "URL to the feature repository. NOT IMPLEMENTED").Default(defaultFeaturesRepoURL).String()

	a.installCmd.cmd = a.app.Command("install", "Install plugins and groovies defined by the 'jplugins.lock'.")
	a.installCmd.lockFile = a.installCmd.cmd.Flag("lock-file", "Full path to the lock file.").Default(lockFileName).String()
	a.installCmd.featureRepoPath = a.installCmd.cmd.Flag("features-repo-path", "Path to a feature repository. "+
		"By default, jplugins store the repo clone in jplugins cache directory.").Default(defaultFeaturesRepoPath).String()
	a.installCmd.featureRepoURL = a.installCmd.cmd.Flag("features-repo-url", "URL to the feature repository. NOT IMPLEMENTED").Default(defaultFeaturesRepoURL).String()
	a.installCmd.jenkinsHomePath = a.installCmd.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()

	// Do not use default git wrapper logOut function.
	git.SetLogFunc(func(msg string) {
		gotrace.Trace(msg)
	})
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

	bError := false
	fileScan := bufio.NewScanner(fd)
	for fileScan.Scan() {
		line := strings.Trim(fileScan.Text(), " \n")
		if gotrace.IsInfoMode() {
			fmt.Printf("== %s ==\n", line)
		}
		lockData.checkElement(line, func(ftype, name, version string) {
			switch ftype {
			case "feature":
				if err = lockData.checkFeature(name); err != nil {
					gotrace.Error("%s", err)
					bError = true
				}
			//case "groovy":
			case "plugin":
				if err := lockData.checkPlugin(name, version, nil); err != nil {
					gotrace.Error("%s", err)
					bError = true
				}
			default:
				gotrace.Warning("feature type '%s' is currently not supported. Ignored.", ftype)
				return
			}
		})
	}

	if bError {
		gotrace.Error("Errors detected. Exiting.")
		return false
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

// readFromJenkins read manifest of each plugins and store information in a.installedPlugins
func (a *jPluginsApp) readFromJenkins(jenkinsHomePath string) (_ bool) {
	pluginsPath := path.Join(jenkinsHomePath, jenkinsHomePluginsPath)

	a.installedElements = make(plugins)

	fEntries, err := ioutil.ReadDir(pluginsPath)

	if err != nil {
		gotrace.Error("Invalid Jenkins home '%s'. %s", pluginsPath, err)
		return
	}

	var fileRE, manifestRE *regexp.Regexp
	fileREDefine := `^(.*)\.[jh]pi*$`
	manifestREDefine := `([\w-]*: )(.*)\n`

	if re, err := regexp.Compile(fileREDefine); err != nil {
		gotrace.Error("Internal error. Regex '%s': %s", fileREDefine, err)
		return
	} else {
		fileRE = re

	}
	if re, err := regexp.Compile(manifestREDefine); err != nil {
		gotrace.Error("Internal error. Regex '%s': %s", fileREDefine, err)
		return
	} else {
		manifestRE = re
	}

	for _, fEntry := range fEntries {
		if fEntry.IsDir() {
			continue
		}

		if fileMatch := fileRE.FindAllStringSubmatch(fEntry.Name(), -1); fileMatch != nil {
			pluginFileName := fileMatch[0][0]
			pluginName := fileMatch[0][1]

			if pluginFileName != "" && pluginName == "" {
				gotrace.Error("Invalid file '%s'. Ignored.", pluginFileName)
				continue
			}

			pluginMetafile := path.Join(pluginsPath, pluginName, "META-INF", "MANIFEST.MF")

			tmpExtract := false
			packagePath := path.Join(pluginsPath, pluginName)
			if _, err := os.Stat(pluginMetafile); err != nil && os.IsNotExist(err) {
				if _, s := utils.RunCmdOutput("unzip", "-q", packagePath+".hpi", "META-INF/MANIFEST.MF", "-d", packagePath); s != 0 {
					gotrace.Error("Unable to extract MANIFEST.MF from plugin package %s", pluginName)
					continue
				}
				tmpExtract = true
			}

			var manifest *elementManifest

			if d, err := ioutil.ReadFile(pluginMetafile); err != nil {
				gotrace.Error("Unable to read file '%s'. %s. Ignored", pluginMetafile, err)
				continue
			} else {
				// Remove DOS format if exist
				data := strings.Replace(string(d), "\r", "", -1)
				// and remove new lines ('\n ')
				data = strings.Replace(data, "\n ", "", -1)
				// and escape "
				data = strings.Replace(data, "\"", "\\\"", -1)
				// and embrace value part with "
				data = manifestRE.ReplaceAllString(data, `$1"$2"`+"\n")

				manifest = new(elementManifest)
				if err := yaml.Unmarshal([]byte(data), &manifest); err != nil {
					gotrace.Error("Unable to read file '%s' as yaml. %s. Ignored", pluginMetafile, err)
					fmt.Print(data)
					continue
				}
				manifest.elementType = "plugin"
			}
			if tmpExtract {
				os.RemoveAll(packagePath)
				tmpExtract = false
			}
			a.installedElements[manifest.Name] = manifest
		}
	}
	return true
}

// readFromSimpleFormat read a simple description file for plugins or groovies.
func (a *jPluginsApp) readFromSimpleFormat(file string) (_ bool) {
	fd, err := os.Open(file)
	if err != nil {
		gotrace.Error("Unable to open file '%s'. %s", file, err)
		return
	}

	defer fd.Close()

	scanFile := bufio.NewScanner(fd)
	a.installedElements = make(plugins)

	for scanFile.Scan() {
		line := scanFile.Text()
		pluginData := new(elementManifest)
		pluginRecord := strings.Split(line, ":")
		if pluginRecord[0] != "plugin" && pluginRecord[0] != "groovy" {
			continue
		}
		pluginData.elementType = pluginRecord[0]
		pluginData.Name = pluginRecord[1]
		if pluginRecord[0] == "plugin" {
			pluginData.Version = pluginRecord[2]
			if refPlugin, found := a.repository.Plugins[pluginData.Name]; !found {
				gotrace.Warning("plugin '%s' is not recognized. Ignored.")
			} else {
				pluginData.LongName = refPlugin.Title
				pluginData.Description = refPlugin.Description
			}
		} else {
			pluginData.commitID = pluginRecord[2]
		}

		a.installedElements[pluginData.Name] = pluginData
	}
	return true
}

func (a *jPluginsApp) printOutVersion(plugins plugins) (_ bool) {
	if a.installedElements == nil {
		return
	}

	pluginsList := make([]string, len(plugins))

	iCount := 0
	for name := range plugins {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	for _, name := range pluginsList {
		fmt.Printf("%s: %s\n", name, plugins[name].Version)
	}
	fmt.Println(iCount, "plugin(s)")
	return true
}

func (a *jPluginsApp) saveVersionAsPreInstalled(jenkinsHomePath string, plugins plugins) (_ bool) {
	if a.installedElements == nil {
		return
	}

	pluginsList := make([]string, len(plugins))

	iCount := 0
	for name := range plugins {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	preInstalledFile := path.Join(jenkinsHomePath, preInstalledFileName)
	piDescriptor, err := os.OpenFile(preInstalledFile, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		gotrace.Error("Unable to create '%s'. %s", preInstalledFile, err)
		return
	}

	defer piDescriptor.Close()

	for _, name := range pluginsList {
		fmt.Fprintf(piDescriptor, "plugin:%s:%s\n", name, plugins[name].Version)
	}
	fmt.Printf("%d plugin(s) saved in '%s'\n", iCount, preInstalledFile)
	return true
}

func (a *jPluginsApp) setVersion() {
	var version string

	if PRERELEASE {
		version = " pre-release V" + VERSION
	} else {
		version = "forjj V" + VERSION
	}

	if build_branch != "master" {
		version += fmt.Sprintf(" branch %s", build_branch)
	}
	if build_tag == "false" {
		version += fmt.Sprintf(" patched - %s - %s", build_date, build_commit)
	}

	a.app.Version(version).Author(AUTHOR)

}
