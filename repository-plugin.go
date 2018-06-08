package main

import (
	"fmt"
	"regexp"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/forjj/utils"
	goversion "github.com/hashicorp/go-version"
)

type repositoryPlugin struct {
	Dependencies   []repositoryDependency
	Name           string
	Version        string
	Title          string
	Description    string `json:"excerpt"`
	versionHistory []versionStruct
	ref            *repository
}

func (p *repositoryPlugin) loadPluginVersionList() []versionStruct {

	if p == nil {
		return nil
	}

	if p.versionHistory != nil {
		return p.versionHistory
	}

	pluginsVersions, err := utils.ReadDocumentFrom(p.ref.repoURLs, p.ref.repoPluginReplace, p.ref.repoPluginSubPaths, p.Name+"/", "text/html")
	if err != nil {
		gotrace.Error("Unable to load '%s'. %s", p.ref.repoFile, err)
		return nil
	}

	versionReString := "'/" + p.ref.repoPluginSubPaths[0] + "/" + p.Name + "/" + `(.*)/.*\.hpi'`
	versionRE, _ := regexp.Compile(versionReString)
	versionList := versionRE.FindAllStringSubmatch(string(pluginsVersions), -1)

	versionHistory := make([]versionStruct, len(versionList))
	iCount := 0
	for _, capturedVersion := range versionList {
		version := versionStruct{}
		err := version.Set(capturedVersion[1])
		if err != nil {
			gotrace.Error("Invalid version string '%s' for plugin '%s'. %s. Ignored", capturedVersion[1], p.Name, err)
			continue
		}
		versionHistory[iCount] = version
		iCount++
	}

	p.versionHistory = versionHistory
	return versionHistory
}

func (p *repositoryPlugin) DetermineVersion(versionConstraints []goversion.Constraints) (version versionStruct, err error) {
	// Search from version history
	version = versionStruct{}
	version.Set(p.Version)
	var history []versionStruct
	for _, constraints := range versionConstraints {
		if history == nil {
			// Check first from central repository data
			if constraints.Check(version.Get()) {
				continue
			}
			// Load the history as we need to go further in the list
			history = p.loadPluginVersionList()
		}

		// The history was loaded... So, check from each elements loaded.
		if len(history) == 0 {
			version = versionStruct{}
			err = fmt.Errorf("%s: No available versions match rule '%s'", p.Name, versionConstraints)
			return
		}
		iCount := 1
		for _, version = range history {
			if !constraints.Check(version.Get()) {
				iCount++
				continue
			}
			history = history[iCount:]
			break
		}
	}
	return
}
