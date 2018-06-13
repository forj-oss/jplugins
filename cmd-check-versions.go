package main

import (
	"path"
	"fmt"

	"github.com/alecthomas/kingpin"
)

type cmdCheckVersions struct {
	cmd              *kingpin.CmdClause
	jenkinsHomePath  *string
	usePreInstalled  *bool
	preInstalledPath *string
}

func (c* cmdCheckVersions)doCheckInstalled() {
	App.repository = NewRepository()
	repo := App.repository
	if !repo.loadFrom() {
		return
	}

	// TODO: Read from lock file
	if *c.usePreInstalled {
		if !App.readFromSimpleFormat(path.Join(*c.preInstalledPath, preInstalledFileName)) {
			return
		}
	} else {
		if !App.readFromJenkins(*c.jenkinsHomePath) {
			return
		}
	}

	updates := repo.compare(App.installedElements)
	updates.displayUpdates()
	fmt.Println(len(repo.Plugins), "plugins/groovies loaded.")
	fmt.Println(len(App.installedElements), "plugins/groovies installed.")

}
