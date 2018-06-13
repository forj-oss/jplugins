package main

import (
	"os"
	"path"

	"github.com/alecthomas/kingpin"

	"github.com/forj-oss/go-git"

	"github.com/forj-oss/forjj-modules/trace"
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

	for _, pathToCheck := range pathsToCheck {
		if info, err := os.Stat(pathToCheck); os.IsNotExist(err) || !info.IsDir() {
			if err != nil {
				gotrace.Error("Unable to install. Invalid Jenkins home dir. %s", err)
			} else {
				gotrace.Error("Unable to install. Invalid Jenkins home dir. %s is not a directory", pathToCheck, err)
			}
			return
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

	for name, element := range App.installedElements {
		switch element.elementType {
		case "groovy":
			groovyObj := newGroovyStatusDetails(name, *c.featureRepoPath)
			groovyObj.newCommit = element.commitID
			if err := groovyObj.installIt(path.Join(*c.jenkinsHomePath, jenkinsHomeGroovyPath)); err != nil {
				gotrace.Error("Installation issue. %s. Ignored.", err)
			}
		case "plugin":
			pluginObj := newPluginsStatusDetails()
			pluginObj.setVersion(element.Version)
			pluginObj.name = element.Name
			if err := pluginObj.installIt(path.Join(*c.jenkinsHomePath, jenkinsHomePluginsPath)); err != nil {
				gotrace.Error("Installation issue. %s. Ignored.", err)
			}
		}
	}
	git.RunInPath(*c.featureRepoPath, func() error {
		git.Do("checkout", savedBranch)
		return nil
	})

}
