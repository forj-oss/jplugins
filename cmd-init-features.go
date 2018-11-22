package main

import (
	"fmt"
	"jplugins/simplefile"
	"os"
	"path"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"

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
	if App.checkJenkinsHome(*c.jenkinsHomePath) {
		if !App.readFromJenkins(*c.jenkinsHomePath) {
			os.Exit(1)
		}
	}
	if !utils.CheckPath(*c.pluginsFeaturePath) {
		gotrace.Error("'%s' is not a valid/accessible features path", *c.pluginsFeaturePath)
		os.Exit(1)
	}

	err := c.saveFeatures()
	if err != nil {
		gotrace.Error("Unable to create the feature file. %s", err)
		os.Exit(1)
	}

}

// saveFeatures
func (c *cmdInitFeatures) saveFeatures() (err error) {
	if utils.CheckFile(*c.pluginsFeaturePath, *c.pluginsFeatureFile) {
		if !*c.replace {
			err = fmt.Errorf("'%s/%s' already exist. Use --force to replace it", *c.pluginsFeaturePath, *c.pluginsFeatureFile)
			return
		}
		gotrace.Info("Replacing '%s/%s'", *c.pluginsFeaturePath, *c.pluginsFeatureFile)
	}

	identified := make(plugins)

	max := 0
	for name, plugin := range App.installedElements {
		identified[name] = plugin
		max++
	}

	for _, plugin := range App.installedElements {
		if plugin.Dependencies == "" {
			continue
		}
		for _, depPluginDetail := range strings.Split(plugin.Dependencies, ",") {
			depPlugin := strings.Split(depPluginDetail, ":")
			if _, found := identified[depPlugin[0]]; found {
				delete(identified, depPlugin[0])
			}
		}
	}

	gotrace.Info("%d/%d features plugin detected. ", len(identified), max)

	featureFile := simplefile.NewSimpleFile(path.Join(*c.pluginsFeaturePath, *c.pluginsFeatureFile), 2)

	for name := range identified {
		featureFile.Add(1, "plugin", name)
	}

	if err = featureFile.WriteSimpleSortedFile(":"); err != nil {
		return err
	}
	gotrace.Info("%s/%s saved.", *c.pluginsFeaturePath, *c.pluginsFeatureFile)

	return
}
