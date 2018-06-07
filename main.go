package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

// App is the application core struct
var App jPluginsApp

func main() {
	App.init()

	switch kingpin.MustParse(App.app.Parse(os.Args[1:])) {
	// Register user
	case App.listInstalled.cmd.FullCommand():
		App.doListInstalled()
	case App.checkVersions.cmd.FullCommand():
		App.doCheckInstalled()
	case App.initCmd.cmd.FullCommand():
		App.doInit()
		/*	case App.update.cmd.FullCommand():
			App.doUpdate()*/
		/*	case App.install.cmd.FullCommand():
			App.doInstall()*/
	}
}
