package main

import (
	"encoding/json"
	"forjj/utils"
	"net/url"

	"github.com/forj-oss/forjj-modules/trace"
)

type repository struct {
	Plugins map[string]repositoryPlugin
	loaded bool
	url    string
}

type repositoryPlugin struct {
	Dependencies []repositoryDependency
	Name        string
	Version     string
	Title       string
	Description string `json:"excerpt"`
}

type repositoryDependency struct {
	Name      string
	Optionnal bool
	Version   string

}

func NewRepository() *repository {
	return new(repository)
}

// loadFrom read an URL file containing the Jenkins updates repository data as json.
func (r *repository) loadFrom(urlString, version, file string) (_ bool) {

	repoUrl, err := url.Parse(urlString)

	if err != nil {
		gotrace.Error("Unable to load '%s'. %s", err)
		return
	}

	var repoData []byte

	repoData, err = utils.ReadDocumentFrom([]*url.URL{repoUrl}, []string{version}, []string{""}, file)
	if err != nil {
		gotrace.Error("Unable to load '%s'. %s", repoUrl.String(), err)
		return
	}

	err = json.Unmarshal(repoData, r)
	if err != nil {
		gotrace.Error("Unable to read '%s'. %s", string(repoData), err)
		return
	}

	return true
}

func (r *repository) compare(plugins plugins) (updates *pluginsStatus) {
	updates = newPluginsStatus()

	updates.compare(plugins, r)
	return
}

func (r *repository) get(name string) (plugin repositoryPlugin, found bool) {
	plugin, found = r.Plugins[name]
	return 
}