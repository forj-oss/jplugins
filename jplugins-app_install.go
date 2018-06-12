package main

import (
	"path"

	"github.com/forj-oss/forjj/git"

	"github.com/forj-oss/forjj-modules/trace"
)

func (a *jPluginsApp) doInstall() {
	a.repository = NewRepository()
	repo := a.repository
	if !repo.loadFrom() {
		return
	}

	// Load the lock file in a.installedPlugins
	if !a.readFromSimpleFormat(*a.installCmd.lockFile) {
		return
	}

	var savedBranch string

	git.RunInPath(*a.installCmd.featureRepoPath, func() error {
		git.Do("stash")
		savedBranch = git.GetCurrentBranch()
		return nil
	})

	for name, element := range a.installedElements {
		switch element.elementType {
		case "groovy":
			groovyObj := newGroovyStatusDetails(name, *a.installCmd.featureRepoPath)
			groovyObj.newCommit = element.commitID
			if err := groovyObj.installIt(path.Join(*a.installCmd.jenkinsHomePath, "init.groovy.d")); err != nil {
				gotrace.Error("Installation issue. %s. Ignored.", err)
			}
		case "plugin":
			pluginObj := newPluginsStatusDetails()
			pluginObj.setVersion(element.Version)
			pluginObj.name = element.Name
			if err := pluginObj.installIt(path.Join(*a.installCmd.jenkinsHomePath, "plugins")); err != nil {
				gotrace.Error("Installation issue. %s. Ignored.", err)
			}
		}
	}
	git.RunInPath(*a.installCmd.featureRepoPath, func() error {
		git.Do("checkout", savedBranch)
		return nil
	})

}
