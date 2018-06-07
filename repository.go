package main

import (
	"encoding/json"
	"github.com/forj-oss/forjj-modules/trace"
	"net/url"
	"forjj/utils"
)

type repository struct {
	Plugins map[string]struct {
		Dependencies []struct{
			Name string
			Optionnal bool
			Version string
		}
		Name string
		Version string
		Title string
	}
	loaded bool
	url string
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

	repoData, err = utils.ReadDocumentFrom([]*url.URL{repoUrl},[]string{version}, []string{""}, file)
	if err != nil {
		gotrace.Error("Unable to load '%s'. %s", err)
		return
	}

	err = json.Unmarshal(repoData, r)
	if err != nil {
		gotrace.Error("Unable to read '%s'. %s", err)
		return
	}

	return true
}

func (r *repository) compare(plugins plugins) (updates *pluginsStatus) {
	updates = newPluginsStatus()

	updates.compare(plugins, r)
	return
}