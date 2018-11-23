package main

import (
	"fmt"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
	core "jplugins/coremgt"

	"jplugins/utils"

	"github.com/alecthomas/kingpin"
)

type cmdInitFeatures struct {
	cmd             *kingpin.CmdClause
	jenkinsHomePath *string

	pluginsFeaturePath *string
	pluginsFeatureFile *string

	replace *bool
}

func (c *cmdInitFeatures) init(parent *kingpin.CmdClause) {
	c.cmd = parent.Command("features", "Initialize the features file from Jenkins home.")
	c.jenkinsHomePath = c.cmd.Flag("jenkins-home", "Path to the Jenkins home.").Default(defaultJenkinsHome).String()
	c.pluginsFeatureFile = c.cmd.Flag("feature-file", "Full path to a feature file. The path must exist.").Default(featureFileName).String()
	c.pluginsFeaturePath = c.cmd.Flag("feature-path", "Feature file name to create.").Default(".").String()

	c.replace = c.cmd.Flag("force", "force to re-create a feature file which already exist.").Bool()

}

func (c *cmdInitFeatures) DoInitFeatures() {
	App.setJenkinsHome(*c.jenkinsHomePath)

	var elements *core.Elements
	if App.checkJenkinsHome() {
		if e, err := App.readFromJenkins() ; err != nil {
			gotrace.Error("%s", err)
			os.Exit(1)
		} else {
			elements = e
		}
	}
	if !utils.CheckPath(*c.pluginsFeaturePath) {
		gotrace.Error("'%s' is not a valid/accessible features path", *c.pluginsFeaturePath)
		os.Exit(1)
	}

	err := c.saveFeatures(elements)
	if err != nil {
		gotrace.Error("Unable to create the feature file. %s", err)
		os.Exit(1)
	}

}

// saveFeatures
func (c *cmdInitFeatures) saveFeatures(elements *core.Elements) (err error) {
	if utils.CheckFile(*c.pluginsFeaturePath, *c.pluginsFeatureFile) {
		if !*c.replace {
			err = fmt.Errorf("'%s/%s' already exist. Use --force to replace it", *c.pluginsFeaturePath, *c.pluginsFeatureFile)
			return
		}
		gotrace.Info("Replacing '%s/%s'", *c.pluginsFeaturePath, *c.pluginsFeatureFile)
	}

	identified := elements.ExtractTopElements()

	gotrace.Info("%d/%d plugin features detected. ", len(identified.GetElements("plugin")), len(elements.GetElements("plugin")))

	if err = identified.WriteSimple(path.Join(*c.pluginsFeaturePath, *c.pluginsFeatureFile), 2); err != nil {
		return
	}
	gotrace.Info("%s/%s saved.", *c.pluginsFeaturePath, *c.pluginsFeatureFile)

	return
}
