package coremgt

import (
	"encoding/json"
	"net/url"

	"github.com/forj-oss/utils"

	"github.com/forj-oss/forjj-modules/trace"
)

type Repository struct {
	Plugins            map[string]*RepositoryPlugin
	loaded             bool
	repoURLs           []*url.URL
	repoReplace        []string
	repoSubPaths       []string
	repoFile           string
	repoPluginReplace  []string
	repoPluginSubPaths []string
}

type RepositoryDependency struct {
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

func NewRepository() (ret *Repository) {
	ret = new(Repository)
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

// loadFromURL read an URL file containing the Jenkins updates repository data as json.
func (r *Repository) LoadFromURL() (_ bool) {
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

	r.SetDefaults()

	return true
}

func (r *Repository) Compare(elements *ElementsType) (updates *PluginsStatus) {
	updates = NewPluginsStatus(elements, r)

	updates.Compare()
	return
}

func (r *Repository) Get(name string) (plugin *RepositoryPlugin, found bool) {
	plugin, found = r.Plugins[name]
	return
}

func (r *Repository) SetDefaults() {
	for _, plugin := range r.Plugins {
		plugin.ref = r
	}
}
