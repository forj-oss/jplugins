package main

import (
	"encoding/json"
	"net/url"

	"github.com/forj-oss/utils"

	"github.com/forj-oss/forjj-modules/trace"
)

type repository struct {
	Plugins            map[string]*repositoryPlugin
	loaded             bool
	repoURLs           []*url.URL
	repoReplace        []string
	repoSubPaths       []string
	repoFile           string
	repoPluginReplace  []string
	repoPluginSubPaths []string
}

type repositoryDependency struct {
	Name      string
	Optionnal bool
	Version   string
}

const (
	JenkinsRepoURL     = "https://updates.jenkins.io"
	JenkinsRepoVersion = "current"
	JenkinsRepoFile    = "update-center.actual.json"
	JenkinsPluginRepo  = "download/plugins"
)

func NewRepository() (ret *repository) {
	ret = new(repository)
	ret.repoSubPaths = []string{JenkinsRepoVersion}
	ret.repoURLs = make([]*url.URL, 1)
	ret.repoURLs[0], _ = url.Parse(JenkinsRepoURL)

	ret.repoReplace = []string{""}
	ret.repoFile = JenkinsRepoFile
	ret.repoPluginReplace = []string{""}
	ret.repoPluginSubPaths = []string{JenkinsPluginRepo}
	return
}

// TODO: Be able to change default repository values

// loadFrom read an URL file containing the Jenkins updates repository data as json.
func (r *repository) loadFrom() (_ bool) {
	repoData, err := utils.ReadDocumentFrom(r.repoURLs, r.repoReplace, r.repoSubPaths, r.repoFile, "")
	if err != nil {
		gotrace.Error("Unable to load '%s'. %s", r.repoFile, err)
		return
	}

	err = json.Unmarshal(repoData, r)
	if err != nil {
		gotrace.Error("Unable to read '%s'. %s", string(repoData), err)
		return
	}

	r.setDefaults()

	return true
}

func (r *repository) compare(plugins plugins) (updates *pluginsStatus) {
	updates = newPluginsStatus(plugins, r)

	updates.compare()
	return
}

func (r *repository) get(name string) (plugin *repositoryPlugin, found bool) {
	plugin, found = r.Plugins[name]
	return
}

func (r *repository) setDefaults() {
	for _, plugin := range r.Plugins {
		plugin.ref = r
	}
}
