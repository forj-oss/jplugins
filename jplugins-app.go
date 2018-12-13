package main

import (
	"bufio"
	"errors"
	"fmt"
	"jplugins/simplefile"
	"os"
	"path"
	"strings"

	core "jplugins/coremgt"
	"jplugins/utils"

	git "github.com/forj-oss/go-git"

	"github.com/alecthomas/kingpin"
	"github.com/forj-oss/forjj-modules/trace"
)

type jPluginsApp struct {
	app           *kingpin.Application
	listInstalled cmdListInstalled
	checkVersions cmdCheckVersions
	initCmd       cmdInit
	installCmd    cmdInstall

	installedElements *core.Plugins
	repository        *core.Repository

	jenkinsHome *core.JenkinsHome
}

const (
	defaultFeaturesRepoName = "jenkins-install-inits"
	defaultFeaturesRepoPath = ".jplugins/repo-cache/" + defaultFeaturesRepoName
	defaultFeaturesRepoURL  = "https://github.com/forj-oss/" + defaultFeaturesRepoName
	defaultJenkinsHome      = "/var/jenkins_home"
	lockFileName            = "jplugins.lock"
	lockBakFileName         = "jplugins.lock.bak"
	featureFileName         = "jplugins.lst"
	preInstalledFileName    = "jplugins-preinstalled.lst"
)

var build_branch, build_commit, build_date, build_tag string

func (a *jPluginsApp) init() {
	gotrace.SetInfo()

	a.app = kingpin.New("jplugins", "Jenkins plugins as Code management tool.")

	a.setVersion()

	a.listInstalled.cmd = a.app.Command("list-installed", "Display Jenkins plugins list of current Jenkins installation.")
	a.listInstalled.jenkinsHomePath = a.listInstalled.cmd.Flag("jenkins-home", "Where Jenkins is installed.").Default(defaultJenkinsHome).String()
	a.listInstalled.preInstalled = a.listInstalled.cmd.Flag("save-pre-installed", "To create the jplugins-preinstalled.lst instead displaying.").Bool()

	a.checkVersions.init()

	a.initCmd.init()

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

func (a *jPluginsApp) setJenkinsHome(jenkinsHomePath string) {
	a.jenkinsHome = core.NewJenkinsHome(jenkinsHomePath)
}

func (a *jPluginsApp) writeLockFile(lockFileName string, lockData *core.PluginsStatus) (_ bool) {

	err := lockData.WriteSimple(lockFileName)
	if err != nil {
		gotrace.Error("Unable to save the lockfile. %s", err)
		return
	}

	gotrace.Info("%s written\n", lockFileName)
	return true
}

func (a *jPluginsApp) readFeatures(featurePath, featureFile, featureURL string, lockData *core.PluginsStatus) (_ bool) {
	gotrace.Trace("Loading constraints...")
	if gotrace.IsDebugMode() {
		fmt.Printf("Reading %s\n--------\n", featureFileName)
	}
	gotrace.Info("Reading %s", featureFileName)
	fd, err := os.Open(featureFile)
	if err != nil {
		gotrace.Error("Unable to read '%s'. %s", featureFileName, err)
		return
	}

	if featurePath != defaultFeaturesRepoPath {
		lockData.SetLocal()
	}
	lockData.SetFeaturesPath(featurePath)
	lockData.SetFeaturesRepoURL(featureURL)

	bError := false
	fileScan := bufio.NewScanner(fd)
	for fileScan.Scan() {
		line := strings.Trim(fileScan.Text(), " \n")
		if gotrace.IsDebugMode() {
			fmt.Printf("== %s ==\n", line)
		}
		lockData.CheckElementLine(line, func(ftype, name, version string) {
			switch ftype {
			case "feature":
				if err = lockData.CheckFeature(name); err != nil {
					gotrace.Error("%s", err)
					bError = true
				}
			//case "groovy":
			case "plugin":
				if err := lockData.CheckPlugin(name, version, nil); err != nil {
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

	if gotrace.IsDebugMode() {
		fmt.Println("--------")
	}
	gotrace.Trace("Identifying version from constraints...")
	lockData.DefinePluginsVersion()

	if !lockData.CheckMinDep() {
		return
	}

	return true
}

// readFeaturesFromSimpleFormat will load a feature file and expand them to get a list of plugins/groovies/... (elements)
func (a *jPluginsApp) readFeaturesFromSimpleFormat(featurePath, featureFile, featureURL string) (elements *core.ElementsType, err error) {
	if gotrace.IsDebugMode() {
		fmt.Println("******** Loading features and build constraints ********")
	}

	elements = core.NewElementsType()

	if featurePath != defaultFeaturesRepoPath {
		elements.SetLocal()
	}
	elements.SetFeaturesPath(featurePath)
	elements.SetFeaturesRepoURL(featureURL)
	elements.SetRepository(a.repository)

	feature := simplefile.NewSimpleFile(featureFile, 3)

	bError := false
	feature.Read(":", func(fields []string) (_ error) {
		_, err := elements.Add(fields...)

		if err != nil {
			gotrace.Error("%s", err)
			bError = true
		}
		return
	})

	if bError {
		err = errors.New("Errors detected. Please review")
		return
	}
	return
}

// checkJenkinsHome verify if the path given exist or not
func (a *jPluginsApp) checkJenkinsHome() (_ bool) {
	if a.jenkinsHome == nil {
		return
	}

	return a.jenkinsHome.IsValid()
}

// readFromJenkins read manifest of each plugins and store information in a.installedPlugins
func (a *jPluginsApp) readFromJenkins() (elements *core.ElementsType, _ error) {
	if a.jenkinsHome == nil {
		return
	}

	return a.jenkinsHome.GetPlugins()
}

// checkSimpleFormatFile simply verify if the file exist.
func (a *jPluginsApp) checkSimpleFormatFile(filepath, file string) (_ bool) {
	return utils.CheckFile(filepath, file)
}

// readFromSimpleFormat read a simple description file for plugins or groovies.
func (a *jPluginsApp) readFromSimpleFormat(filepath, fileName string) (elements *core.ElementsType, _ error) {
	file := path.Join(filepath, fileName)
	elements = core.NewElementsType()

	elements.AddSupport("plugin", "groovy")
	elements.AddSupportContext("groovy", "noMoreContext", "true")
	elements.SetRepository(a.repository)

	err := elements.Read(file, 3)
	if err != nil {
		return nil, fmt.Errorf("Unable to open file simple file format'%s'. %s", file, err)
	}

	return
}

// printOutVersion display the list of plugins given
func (a *jPluginsApp) printOutVersion(plugins *core.ElementsType) (_ bool) {
	if plugins == nil {
		return
	}

	plugins.PrintOut(func(element core.Element) {
		version, _ := element.GetVersion()
		fmt.Printf("%s: %s\n", element.Name(), version)

	})

	fmt.Printf("\n%d plugin(s)\n", plugins.Length())
	return true
}

// saveVersionAsPreInstalled store the list of plugins in the jenkinsHomePath
func (a *jPluginsApp) saveVersionAsPreInstalled(jenkinsHomePath string, plugins *core.ElementsType) (_ bool) {
	if plugins == nil {
		return
	}

	preInstalledFile := path.Join(jenkinsHomePath, preInstalledFileName)
	piDescriptor, err := os.OpenFile(preInstalledFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		gotrace.Error("Unable to create '%s'. %s", preInstalledFile, err)
		return
	}

	defer piDescriptor.Close()

	plugins.PrintOut(func(element core.Element) {
		version, _ := element.GetVersion()
		fmt.Fprintf(piDescriptor, "plugin:%s:%s\n", element.Name(), version)
	})

	fmt.Printf("%d plugin(s) saved in '%s'\n", plugins.Length(), preInstalledFile)
	return true
}

// setVersion define the current jplugins version.
func (a *jPluginsApp) setVersion() {
	version := "jplugins"

	if PRERELEASE {
		version += " pre-release V" + VERSION
	} else if build_tag == "false" {
		version += " pre-version V" + VERSION
	} else {
		version += " V" + VERSION
	}

	if build_branch != "master" && build_branch != "HEAD" {
		version += fmt.Sprintf(" branch %s", build_branch)
	}
	if build_tag == "false" {
		version += fmt.Sprintf(" - %s - %s", build_date, build_commit)
	}

	a.app.Version(version).Author(AUTHOR)

}
