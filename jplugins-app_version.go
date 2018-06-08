package main

import (
	"fmt"
)

var build_branch, build_commit, build_date, build_tag string

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
