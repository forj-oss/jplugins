package coremgt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
	goversion "github.com/hashicorp/go-version"
)

type pluginsStatusDetails struct {
	name          string
	title         string
	oldVersion    VersionStruct
	newVersion    VersionStruct
	minDepVersion VersionStruct
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
func (sd *pluginsStatusDetails) initFromRef(oldVersion VersionStruct, plugin *RepositoryPlugin) *pluginsStatusDetails {
	if sd == nil {
		return nil
	}

	sd.name = plugin.Name
	sd.title = plugin.Title
	version := VersionStruct{}
	var err error

	err = version.Set(plugin.Version)
	if err != nil {
		gotrace.Error("New version '%s' invalid. %s", plugin.Version, err)
		return nil
	}

	sd.newVersion = version
	sd.oldVersion = oldVersion

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

	newVersion := VersionStruct{}
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
	depVersion := VersionStruct{}
	depVersion.Set(version)

	if minVersion := sd.minDepVersion.Get(); minVersion == nil || minVersion.LessThan(depVersion.Get()) {
		sd.minDepVersion = depVersion
	}
}

func (sd *pluginsStatusDetails) checkMinimumVersionDep(version string, parentPlugin *pluginsStatusDetails) {
	if sd == nil {
		return
	}
	depVersion := VersionStruct{}
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

func (sd *pluginsStatusDetails) initAsObsolete(element Element) *pluginsStatusDetails {
	if sd == nil {
		return nil
	}

	version, err := element.GetVersion()
	if err != nil {
		gotrace.Error("New version '%s' invalid. %s", version, err)
		return nil
	}
	sd.newVersion = version

	sd.oldVersion = version
	sd.name = element.Name()
	sd.title = element.(*Plugin).LongName

	return sd
}

func (sd *pluginsStatusDetails) setIsLatest() {
	if sd == nil {
		return
	}

	sd.latest = true
}

func (sd *pluginsStatusDetails) installIt(destPath string) (err error) {
	var resp *http.Response
	pluginURL := JenkinsRepoURL + "/" + JenkinsPluginRepo + "/" + sd.name + "/" + sd.newVersion.String() + "/" + path.Base(sd.name) + ".hpi"
	destFile := path.Join(destPath, path.Base(sd.name)+".hpi")

	if resp, err = http.Get(pluginURL); err != nil {
		err = fmt.Errorf("Unable to read '%s'. %s", pluginURL, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return fmt.Errorf("File %s not found", pluginURL)
	}

	var destfd *os.File
	destfd, err = os.Create(destFile)
	if err != nil {
		return err
	}
	defer destfd.Close()

	_, err = io.Copy(destfd, resp.Body)
	if err != nil {
		return fmt.Errorf("Unable to copy %s to %s. %s", pluginURL, destFile, err)
	}

	gotrace.Trace("Copied: %s => %s", pluginURL, destFile)

	return nil
}
