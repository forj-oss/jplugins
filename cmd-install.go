package main

import (
	"os"
	core "jplugins/coremgt"

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

func (c *cmdInstall) doInstall() {
	App.repository = core.NewRepository()
	repo := App.repository
	if !repo.LoadFromURL() {
		os.Exit(1)
	}

	var elements *core.ElementsType

	// Load the lock file in App.installedPlugins
	if e, err := App.readFromSimpleFormat("", *c.lockFile) ; err != nil {
		gotrace.Error("%s", err)
		os.Exit(1)
	} else {
		elements = e
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

	jenkinsHome := core.NewJenkinsHome(*c.jenkinsHomePath)

	if err := jenkinsHome.Install(elements, *c.featureRepoPath) ; err != nil {
		gotrace.Error("%s. Process aborted.", err)
		os.Exit(1)
	}
}
