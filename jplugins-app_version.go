package main

import (
	"fmt"
)

var build_branch, build_commit, build_date, build_tag string

const (
	// VERSION application
	VERSION = "0.0.1"
	// PRERELEASE = true if the version exposed is a pre-release
	PRERELEASE = true
	// AUTHOR is the project maintainer
	AUTHOR = "Christophe Larsonneur <clarsonneur@gmail.com>"
)

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
