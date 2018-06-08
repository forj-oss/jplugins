package main

import (
	"github.com/forj-oss/forjj-modules/trace"
	goversion "github.com/hashicorp/go-version"
)

type pluginsStatusDetails struct {
	name         string
	title        string
	oldVersion   versionStruct
	newVersion   versionStruct
	rules        []goversion.Constraints
	preInstalled bool
}

func newPluginsStatusDetails() (ret *pluginsStatusDetails) {
	ret = new(pluginsStatusDetails)
	ret.rules = make([]goversion.Constraints, 0, 5)
	return
}

// setFromRef creates the version from reference and add current version as old.
// newVersion is by default set to latest from ref.
func (sd *pluginsStatusDetails) initFromRef(oldVersion string, plugin *repositoryPlugin) *pluginsStatusDetails {
	if sd == nil {
		return nil
	}

	sd.name = plugin.Name
	sd.title = plugin.Title
	version := versionStruct{}
	var err error

	err = version.Set(plugin.Version)
	if err != nil {
		gotrace.Error("New version '%s' invalid. %s", plugin.Version, err)
		return nil
	}

	sd.newVersion = version

	err = version.Set(oldVersion)
	if err != nil {
		gotrace.Error("New version '%s' invalid. %s", plugin.Version, err)
		return nil
	}
	sd.oldVersion = version

	return sd
}

func (sd *pluginsStatusDetails) setAsPreInstalled() *pluginsStatusDetails {
	if sd == nil {
		return nil
	}
	sd.preInstalled = true
	return sd
}

func (sd *pluginsStatusDetails) setVersion(version string) *pluginsStatusDetails {
	if sd == nil {
		return nil
	}

	newVersion := versionStruct{}
	var err error

	err = newVersion.Set(version)
	if err != nil {
		gotrace.Error("New version '%s' invalid. %s", version, err)
		return nil
	}

	sd.newVersion = newVersion
	return sd
}

func (sd *pluginsStatusDetails) addConstraint(constraints goversion.Constraints) *pluginsStatusDetails {
	if sd == nil {
		return nil
	}
	sd.rules = append(sd.rules, constraints)
	return sd
}

func (sd *pluginsStatusDetails) initAsObsolete(plugin *pluginManifest) *pluginsStatusDetails {
	if sd == nil {
		return nil
	}

	version := versionStruct{}
	sd.newVersion = version

	var err error

	err = version.Set(plugin.Version)
	if err != nil {
		gotrace.Error("New version '%s' invalid. %s", plugin.Version, err)
		return nil
	}

	sd.oldVersion = version
	sd.name = plugin.Name
	sd.title = plugin.LongName

	return sd
}