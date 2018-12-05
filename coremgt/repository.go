package coremgt

import (
	"encoding/json"
	"net/url"
	"sort"

	"github.com/forj-oss/utils"

	"github.com/forj-oss/forjj-modules/trace"

	goversion "github.com/hashicorp/go-version"
)

type Repository struct {
	RepositoryPlugins  // Loaded from json with LoadFromURL and JenkinsRepoFile
	historyPlugins     RepositoryPluginsHistory
	loaded             bool
	repoURLs           []*url.URL
	repoReplace        []string
	repoSubPaths       []string
	repoFile           string
	repoHistoryFile    string
	repoPluginReplace  []string
	repoPluginSubPaths []string
}

type RepositoryDependency struct {
	Name      string
	Optionnal bool
	Version   string
}

type RepositoryPlugins struct {
	Plugins map[string]*RepositoryPlugin
}

// RepositoryHistory is the plugin-versions.json data representation
type RepositoryPluginsHistory struct {
	Plugins map[string]map[string]*RepositoryPlugin
}

const (
	JenkinsRepoURL     = "https://updates.jenkins.io"
	JenkinsRepoVersion = "current"
	JenkinsRepoFile    = "update-center.actual.json"
	JenkinsHistoryFile = "plugin-versions.json" // TODO: Use this to load it so we can identify better version from constraints.
	JenkinsPluginRepo  = "download/plugins"
)

func NewRepository() (ret *Repository) {
	ret = new(Repository)
	ret.repoSubPaths = []string{JenkinsRepoVersion}
	ret.repoURLs = make([]*url.URL, 1)
	ret.repoURLs[0], _ = url.Parse(JenkinsRepoURL)

	ret.repoReplace = []string{""}
	ret.repoFile = JenkinsRepoFile
	ret.repoHistoryFile = JenkinsHistoryFile
	ret.repoPluginReplace = []string{""}
	ret.repoPluginSubPaths = []string{JenkinsPluginRepo}
	return
}

// TODO: Be able to change default repository values

// LoadFromURL read an URL file containing the Jenkins updates repository data as json.
func (r *Repository) LoadFromURL() (_ bool) {
	gotrace.Info("1/2 Loading repositories... %s", r.repoFile)
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

	gotrace.Info("2/2 Loading repositories... %s", r.repoHistoryFile)
	repoData, err = utils.ReadDocumentFrom(r.repoURLs, r.repoReplace, r.repoSubPaths, r.repoHistoryFile, "")
	if err != nil {
		gotrace.Error("Unable to load '%s'. %s", r.repoHistoryFile, err)
		return
	}

	err = json.Unmarshal(repoData, &r.historyPlugins)
	if err != nil {
		gotrace.Error("Unable to read json data from '%s'. %s", r.repoHistoryFile, err)
		return
	}

	gotrace.Info("Repositories loaded.")
	r.setDefaults()

	return true
}

// Compare creates a PluginsStatus which store old and new version of each elements.
func (r *Repository) Compare(elements *ElementsType) (updates *PluginsStatus) {
	updates = NewPluginsStatus(elements, r)

	updates.Compare()
	return
}

// Get return the plugin requested (name with or without version) as described by Jenkins updates.
//
// If the version is defined but not found, it will return the closest higher version matching this unknown version
func (r *Repository) Get(pluginRequested ...string) (plugin *RepositoryPlugin, found bool) {
	if len(pluginRequested) < 1 {
		return
	}
	name := pluginRequested[0]
	version := "latest"
	if len(pluginRequested) >= 2 && pluginRequested[1] != "" {
		version = pluginRequested[1]
	}
	if version == "latest" {
		plugin, found = r.Plugins[name]
	} else {
		if pluginVersions, foundPlugin := r.historyPlugins.Plugins[name]; foundPlugin {
			plugin, found = pluginVersions[version]
			if !found { // We need to get the upper version from the history list for the given unknown version.
				versions := make([]*goversion.Version, 0, len(pluginVersions))
				for version := range pluginVersions {
					versionToAdd, _ := goversion.NewVersion(version)
					versions = append(versions, versionToAdd)
				}
				sort.Sort(goversion.Collection(versions))
				versionToCompare, _ := goversion.NewConstraint(">=" + version)
				for _, versionToCheck := range versions {
					if versionToCompare.Check(versionToCheck) {
						version = versionToCheck.Original()
						plugin = pluginVersions[version]
						break
					}
				}
			}
			if plugin == nil {
				return nil, false
			}
			pluginInfo := r.Plugins[name]
			plugin.Description = pluginInfo.Description
			plugin.Title = pluginInfo.Title
		}
	}
	return
}

// GetVersions return the list of a plugin version for a given plugin
func (r *Repository) GetVersions(name string) (_ map[string]*RepositoryPlugin) {
	if list, found := r.historyPlugins.Plugins[name]; found {
		return list
	}
	return
}

// GetOrderedVersions return the list of a plugin version for a given plugin ordered from latest to oldest.
func (r *Repository) GetOrderedVersions(name string) (ret goversion.Collection) {

	refVersions := r.GetVersions(name)
	versions := make([]*goversion.Version, 0, len(refVersions))
	for version := range r.GetVersions(name) {
		v, _ := goversion.NewVersion(version)
		if v != nil {
			versions = append(versions, v)
		}
	}

	ret = goversion.Collection(versions)

	sort.Sort(ret)

	// Reversing the order
	for left, right := 0, len(ret)-1; left < right; left, right = left+1, right-1 {
		ret[left], ret[right] = ret[right], ret[left]
	}

	return
}

func (r *Repository) setDefaults() {
	for _, plugin := range r.Plugins {
		plugin.ref = r
	}
	for _, pluginVersions := range r.historyPlugins.Plugins {
		for _, pluginVersion := range pluginVersions {
			pluginVersion.ref = r
		}
	}
}
