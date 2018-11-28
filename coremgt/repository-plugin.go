package coremgt

import (
	"fmt"
	"regexp"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/utils"
	goversion "github.com/hashicorp/go-version"
)

type RepositoryPlugin struct {
	Dependencies   []RepositoryDependency
	Name           string
	Version        string
	Title          string
	Description    string `json:"excerpt"`
	JenkinsVersion string `json:"requiredCore"`
	versionHistory []VersionStruct
	ref            *Repository
}

func (p *RepositoryPlugin) loadPluginVersionList() []VersionStruct {

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

	versionHistory := make([]VersionStruct, len(versionList))
	iCount := 0
	for _, capturedVersion := range versionList {
		version := VersionStruct{}
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

// DetermineVersion select the appropriate version of a plugin depending on constraints.
func (p *RepositoryPlugin) DetermineVersion(versionConstraints map[string]goversion.Constraints) (version VersionStruct, latest bool, err error) {
	// Search from version history
	version = VersionStruct{}
	version.Set(p.Version)
	var history []VersionStruct
	gotrace.Trace("Determining version for '%s'. %d constraints to verify", p.Name, len(versionConstraints))
	latest = true
	for _, constraints := range versionConstraints {
		gotrace.Trace("Constraint to check: '%s'", constraints)
		if history == nil {
			// Check first from central repository data
			if constraints.Check(version.Get()) {
				gotrace.Trace("0: %s - %s : OK", version.Get(), constraints)
				continue
			}
			gotrace.Trace("0: %s - %s : NO", version.Get(), constraints)
			gotrace.Trace("Getting more versions from history...")
			// Load the history as we need to go further in the list
			history = p.loadPluginVersionList()
			history = history[1:]
		}

		latest = false
		// The history was loaded... So, check from each elements loaded.
		if len(history) == 0 {
			version = VersionStruct{}
			err = fmt.Errorf("%s: No available versions match rule '%s'", p.Name, versionConstraints)
			return
		}
		iCount := 1
		for _, version = range history {
			if !constraints.Check(version.Get()) {
				gotrace.Trace("%d: %s - %s : NO", iCount, version.Get(), constraints)
				iCount++
				continue
			}
			gotrace.Trace("%d: %s - %s : OK", iCount, version.Get(), constraints)
			history = history[iCount-1:]
			break
		}
		if !constraints.Check(version.Get()) {
			gotrace.Error("%s: Failed to find a version that respect %s. You can to review your feature.lst and dependencies")
		}
	}
	return
}

func (p *RepositoryPlugin) GetVersion() (ret VersionStruct, err error) {
	err = ret.Set(p.Version)
	return
}
