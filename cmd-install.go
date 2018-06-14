package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"

	"github.com/forj-oss/forjj-modules/trace"
	git "github.com/forj-oss/go-git"
)

type cmdInstall struct {
	cmd             *kingpin.CmdClause
	lockFile        *string
	featureRepoPath *string
	featureRepoURL  *string
	jenkinsHomePath *string
}

const (
	jenkinsHomeGroovyPath  = "init.groovy.d"
	jenkinsHomePluginsPath = "plugins"
)

func (c *cmdInstall) doInstall() {
	pathsToCheck := []string{
		path.Join(*c.jenkinsHomePath, jenkinsHomeGroovyPath),
		path.Join(*c.jenkinsHomePath, jenkinsHomePluginsPath),
	}

	filesToCleanUp := make([]*regexp.Regexp, 2)
	filesToCleanUp[0], _ = regexp.Compile(`^.*\.groovy$`)
	filesToCleanUp[1], _ = regexp.Compile(`^.*\.[jh]pi$`)

	for index, pathToCheck := range pathsToCheck {
		if info, err := os.Stat(pathToCheck); os.IsNotExist(err) || !info.IsDir() {
			if err != nil {
				gotrace.Error("Unable to install. Invalid Jenkins home dir. %s", err)
			} else {
				gotrace.Error("Unable to install. Invalid Jenkins home dir. %s is not a directory", pathToCheck, err)
			}
			return
		}

		cleanupDir, _ := ioutil.ReadDir(pathToCheck)
		for _, element := range cleanupDir {
			if fileName := element.Name(); filesToCleanUp[index].MatchString(fileName) {
				if err := os.Remove(path.Join(pathToCheck, fileName)); err != nil {
					gotrace.Error("Cleanup failure. %s", err)
					os.Exit(1)
				}
			}
		}
	}

	App.repository = NewRepository()
	repo := App.repository
	if !repo.loadFrom() {
		return
	}

	// Load the lock file in App.installedPlugins
	if !App.readFromSimpleFormat(*c.lockFile) {
		return
	}

	var savedBranch string

	git.RunInPath(*c.featureRepoPath, func() error {
		git.Do("stash")
		savedBranch = git.GetCurrentBranch()
		return nil
	})
	defer git.RunInPath(*c.featureRepoPath, func() error {
		git.Do("checkout", savedBranch)
		return nil
	})

	iCountGroovy := 0
	iCountPlugin := 0
	iCountError := 0
	iCountObsolete := 0

	elementsList := make([]string, len(App.installedElements))

	iCount := 0
	iMaxName := 0
	for name, element := range App.installedElements {
		value := element.elementType + ":" + name
		elementsList[iCount] = value
		if size := len(value); size > iMaxName {
			iMaxName = size
		}
		iCount++
	}

	nameFormat := "- %-" + strconv.Itoa(iMaxName) + "s ... "

	iCount = 0
	sort.Strings(elementsList)

	for _, displayName := range elementsList {
		names := strings.Split(displayName, ":")
		name := names[1]
		element := App.installedElements[name]

		switch element.elementType {
		case "groovy":
			if element.commitID == "" {
				if !gotrace.IsDebugMode() {
					fmt.Printf(nameFormat, displayName)
					fmt.Print("obsolete - removed if found.")
				}
				iCountObsolete++
				continue
			}

			groovyObj := newGroovyStatusDetails(name, *c.featureRepoPath)
			groovyObj.newCommit = element.commitID
			if !gotrace.IsDebugMode() {
				fmt.Printf(nameFormat, displayName)
			}
			if err := groovyObj.installIt(path.Join(*c.jenkinsHomePath, jenkinsHomeGroovyPath)); err != nil {
				gotrace.Error("Installation issue. %s. Ignored.", err)
				iCountError++
			} else {
				iCountGroovy++
				iCount++
				if !gotrace.IsDebugMode() {
					fmt.Printf(" installed - %s\n", element.commitID)
				}
			}

		case "plugin":
			if element.Version == "" {
				if !gotrace.IsDebugMode() {
					fmt.Printf(nameFormat, displayName)
					fmt.Print("obsolete - removed if found.")
				}
				iCountObsolete++
				continue
			}

			pluginObj := newPluginsStatusDetails()
			pluginObj.setVersion(element.Version)
			pluginObj.name = element.Name
			if !gotrace.IsDebugMode() {
				fmt.Printf(nameFormat, displayName)
			}
			if err := pluginObj.installIt(path.Join(*c.jenkinsHomePath, jenkinsHomePluginsPath)); err != nil {
				gotrace.Error("Installation issue. %s. Ignored.", err)
				iCountError++
			} else {
				iCountPlugin++
				iCount++
				if !gotrace.IsDebugMode() {
					fmt.Printf(" installed - %s\n", element.Version)
				}
			}
		}
	}

	gotrace.Info("Total %d installed: %d plugins, %d groovies. %d obsoleted. %d error(s) found.\n", iCount, iCountPlugin, iCountGroovy, iCountObsolete, iCountError)
	if iCountError > 0 {
		gotrace.Error("Errors detected. Process aborted.")
		os.Exit(1)
	}
}
