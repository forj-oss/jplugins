package main

import (
	"path"
)

func (a *jPluginsApp) doInstall() {
	// Load the lock file in a.installedPlugins
	if !a.readFromSimpleFormat(*a.installCmd.lockFile) {
		return
	}
	for name, element := range a.installedElements {
		switch element.elementType {
		case "groovy":
			groovyObj := newGroovyStatusDetails(name, *a.installCmd.featureRepoPath)
			groovyObj.installIt(path.Join(*a.installCmd.jenkinsHomePath, "init.groovy.d"))
		case "plugin":
			pluginObj := newPluginsStatusDetails()
			pluginObj.setVersion(element.Version)
			pluginObj.name = element.Name
			pluginObj.installIt(path.Join(*a.installCmd.jenkinsHomePath, "plugins"))
		}
	}
}
