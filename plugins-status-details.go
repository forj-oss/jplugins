package main

import (
	"fmt"

	"github.com/forj-oss/forjj-modules/trace"
	goversion "github.com/hashicorp/go-version"
)

type pluginsStatusDetails struct {
	name          string
	title         string
	oldVersion    versionStruct
	newVersion    versionStruct
	minDepVersion versionStruct
	minDepName    string
	latest        bool
	rules         map[string]goversion.Constraints
	preInstalled  bool
}

func newPluginsStatusDetails() (ret *pluginsStatusDetails) {
	ret = new(pluginsStatusDetails)
	ret.rules = make(map[string]goversion.Constraints)
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

func (sd *pluginsStatusDetails) setMinimumVersionDep(version string) {
	if sd == nil {
		return
	}
	depVersion := versionStruct{}
	depVersion.Set(version)

	if minVersion := sd.minDepVersion.Get(); minVersion == nil || minVersion.LessThan(depVersion.Get()) {
		sd.minDepVersion = depVersion
	}
}

func (sd *pluginsStatusDetails) checkMinimumVersionDep(version string, parentPlugin *pluginsStatusDetails) {
	if sd == nil {
		return
	}
	depVersion := versionStruct{}
	depVersion.Set(version)

	if sd.newVersion.Get().LessThan(depVersion.Get()) {
		sd.minDepName += fmt.Sprintf("%s:%s requires %s:%s, ", parentPlugin.name, parentPlugin.newVersion, sd.name, version)
	}
}

func (sd *pluginsStatusDetails) addConstraint(constraintsGiven string) *pluginsStatusDetails {
	if sd == nil {
		return nil
	}

	constraints, err := goversion.NewConstraint(constraintsGiven)
	if err != nil {
		gotrace.Error("Version constraints are invalid. %s. Ignored", err)
		return nil
	}

	sd.rules[constraints.String()] = constraints
	return sd
}

func (sd *pluginsStatusDetails) initAsObsolete(plugin *elementManifest) *pluginsStatusDetails {
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

func (sd *pluginsStatusDetails) setIsLatest() {
	if sd == nil {
		return
	}

	sd.latest = true
}

func (sd *pluginsStatusDetails) installIt(destPath string) (error) {
	return nil
}