package main

import (
	"bufio"
	"fmt"
	"forjj/utils"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/forj-oss/forjj/git"

	goversion "github.com/hashicorp/go-version"

	"github.com/forj-oss/forjj-modules/trace"
)

type pluginsStatus struct {
	plugins   map[string]*pluginsStatusDetails
	installed plugins
	ref       *repository
	repoPath  string
	repoURL   []*url.URL
	useLocal  bool
}

func newPluginsStatus(installed plugins, ref *repository) (pluginsCompared *pluginsStatus) {
	pluginsCompared = new(pluginsStatus)
	pluginsCompared.plugins = make(map[string]*pluginsStatusDetails)
	pluginsCompared.installed = installed
	pluginsCompared.ref = ref
	pluginsCompared.repoURL = make([]*url.URL, 0, 3)
	return
}

// setLocal set the useLocal to true
// When set to true, jplugin do not clone a remote repo URL to store on cache
func (s *pluginsStatus) setLocal() {
	if s == nil {
		return
	}
	s.useLocal = true
}

func (s *pluginsStatus) setFeaturesRepoURL(repoURL string) error {
	if s == nil {
		return nil
	}

	if repoURLObject, err := url.Parse(repoURL); err != nil {
		return fmt.Errorf("Invalid feature repository URL. %s", repoURL)
	} else {
		s.repoURL = append(s.repoURL, repoURLObject)
	}
	return nil
}

func (s *pluginsStatus) setFeaturesPath(repoPath string) error {
	if s == nil {
		return nil
	}

	if p, err := utils.Abs(repoPath); err != nil {
		return fmt.Errorf("Invalid feature repository path. %s", repoPath)
	} else {
		s.repoPath = p
	}
	return nil
}

func (s *pluginsStatus) compare() {
	installed := s.installed
	ref := s.ref
	for name, plugin := range installed {
		refPlugin, found := ref.get(name)
		if !found {
			s.obsolete(plugin)
			continue
		}
		if plugin.Version != refPlugin.Version {
			s.add(plugin.Version, refPlugin)
		}

		for _, dep := range refPlugin.Dependencies {
			if dep.Optionnal {
				return
			}
			if _, found = installed[dep.Name]; !found {

				if p, found := ref.get(dep.Name); found {
					s.add("new", p)
				} else {
					gotrace.Trace("Internal repo error: From '%s', dependency '%s' has not been found.", name, dep.Name)
					continue
				}
			}
		}
	}
}

func (s *pluginsStatus) addConstraints(constraints goversion.Constraints, plugin *repositoryPlugin) (_ bool) {
	pluginLock, found := s.plugins[plugin.Name]
	if !found {
		return
	}

	pluginLock.addConstraint(constraints)
	return true
}

// chooseNewVersion change the default new version of a locked plugin
func (s *pluginsStatus) chooseNewVersion(name, version string) (_ bool) {
	pluginLock, found := s.plugins[name]
	if !found {
		return
	}

	pluginLock.setVersion(version)
	return true
}

// set do add/update a plugin version to the pluginsStatus structure
// The version given is the current version use.
func (s *pluginsStatus) add(version string, pluginRef *repositoryPlugin) (_ bool) {
	_, found := s.plugins[pluginRef.Name]

	if found {
		return
	}

	s.plugins[pluginRef.Name] = newPluginsStatusDetails().initFromRef(version, pluginRef)

	return true
}

func (s *pluginsStatus) obsolete(plugin *pluginManifest) {
	_, found := s.plugins[plugin.Name]

	if found {
		return
	}

	s.plugins[plugin.Name] = newPluginsStatusDetails().initAsObsolete(plugin)
}

func (s *pluginsStatus) displayUpdates() (_ bool) {
	if s == nil {
		return
	}

	if len(s.plugins) == 0 {
		fmt.Print("No updates detected.")
		return true
	}

	pluginsList := make([]string, len(s.plugins))
	pluginsDetails := make(map[string]*pluginsStatusDetails)
	iCount := 0
	iMaxTitle := 0
	for _, plugin := range s.plugins {
		pluginsList[iCount] = plugin.title
		pluginsDetails[plugin.title] = plugin
		if val := len(plugin.title) + len(plugin.name); val > iMaxTitle {
			iMaxTitle = val
		}
		iCount++
	}

	sort.Strings(pluginsList)

	iCountUpdated := 0
	iCountNew := 0
	for _, title := range pluginsList {
		plugin := pluginsDetails[title]
		if old := plugin.oldVersion.String(); old == plugin.newVersion.String() {
			fmt.Printf("%-"+strconv.Itoa(iMaxTitle+3)+"s : %-10s => No update\n", title+" ("+plugin.name+")", plugin.oldVersion)
		} else {
			iCountUpdated++
			if old == "new" {
				iCountNew++
			}
			fmt.Printf("%-"+strconv.Itoa(iMaxTitle+3)+"s : %-10s => %s\n", title+" ("+plugin.name+")", plugin.oldVersion, plugin.newVersion)
		}

	}

	fmt.Printf("\nFound %d/%d plugin(s) updates available. %d are new.\n", iCountUpdated, iCount, iCountNew)

	return true
}

func (s *pluginsStatus) importInstalled(pluginsData plugins) {
	for name, plugin := range pluginsData {
		if _, found := s.plugins[name]; found {
			gotrace.Warning("plugin '%s' is duplicated. Ignored. FYI, it is better to extract pre-installed version list from a fresh jenkins installation, like with Docker.", name)
			continue
		}

		pluginRef, found := s.ref.get(name)
		if !found {
			s.plugins[name] = newPluginsStatusDetails().initAsObsolete(plugin)
		} else {
			constraints, err := goversion.NewConstraint(">=" + plugin.Version)

			if err != nil {
				gotrace.Warning("plugin '%s' has a malformed version. Ignored", name)
			}
			s.plugins[name] = newPluginsStatusDetails().
				initFromRef(plugin.Version, pluginRef).
				setAsPreInstalled().
				addConstraint(constraints)
		}
	}
}

func (s *pluginsStatus) checkElement(line string, split func(string, string, string)) {
	var ftype, fname string
	var fversion string

	if line == "" || line[0] == '#' {
		return
	}

	fields := strings.Split(line, ":")
	switch len(fields) {
	case 1:
		gotrace.Warning("Line %s: Line format is incorrect. It should be <'plugin'|'feature'>:<plugin Name>[:<version>]", line)
		return
	case 2:
		ftype = strings.Trim(fields[0], " ")
		fname = strings.Trim(fields[1], " ")
	default:
		ftype = strings.Trim(fields[0], " ")
		fname = strings.Trim(fields[1], " ")
		fversion = strings.Trim(fields[2], " ")
	}

	split(ftype, fname, fversion)

}

func (s *pluginsStatus) checkFeature(name string) (_ bool) {
	if s == nil {
		return
	}
	if !s.useLocal {
		gotrace.Error("Git clone of repository not currently implemented.")
		return
	}

	if err := git.RunInPath(s.repoPath, func() error {
		if git.Do("rev-parse", "--git-dir") != 0 {
			return fmt.Errorf("Not a valid GIT repository.")
		}
		return nil
	}); err != nil {
		gotrace.Error("Issue with '%s', %s", s.repoPath, err)
		return
	}

	featureFile := path.Join(s.repoPath, name, name+".desc")
	fd, err := os.Open(featureFile)
	if err != nil {
		gotrace.Error("Unable to read feature file '%s'. %s", featureFile, err)
		return
	}
	defer fd.Close()

	fileScan := bufio.NewScanner(fd)
	for fileScan.Scan() {
		line := strings.Trim(fileScan.Text(), " \n")
		if gotrace.IsInfoMode() {
			fmt.Printf("== >> %s ==\n", line)
		}
		s.checkElement(line, func(ftype, name, version string) {
			switch ftype {
			//case "groovy":
			case "plugin":
				s.checkPlugin(name, version)
			default:
				gotrace.Warning("feature type '%s' is currently not supported. Ignored.", ftype)
				return
			}

		})
	}
	return true
}

func (s *pluginsStatus) checkPlugin(name, versionConstraints string) {
	var constraints goversion.Constraints

	refPlugin, found := s.ref.get(name)
	if !found {
		gotrace.Error("Plugin '%s' not found in the public repository. Ignored.", name)
		return
	}

	if _, found := s.plugins[name]; !found {
		if !s.add("new", refPlugin) {
			return
		}
		gotrace.Info("New plugin '%s' identified.", name)
	}

	if versionConstraints != "" {
		if v, err := goversion.NewConstraint(versionConstraints); err == nil {
			constraints = v
		} else {
			gotrace.Error("Version constraints are invalid. %s. Ignored", err)
			return
		}

		s.addConstraints(constraints, refPlugin)
	}

	for _, dep := range refPlugin.Dependencies {
		if dep.Optionnal {
			continue
		}
		s.checkPlugin(dep.Name, ">="+dep.Version)
	}

}

func (s *pluginsStatus) definePluginsVersion() (_ bool) {
	for name, plugin := range s.plugins {
		refPlugin, found := s.ref.get(name)
		if !found {
			gotrace.Error("Plugin '%s' not found in the public repository. Ignored.", name)
			return
		}
		if foundVersion, err := refPlugin.DetermineVersion(plugin.rules); err != nil {
			gotrace.Error("Unable to find a version for plugin '%s' which respect all rules. %s. Please fix it", name, err)
		} else {
			plugin.setVersion(foundVersion.String())
			if plugin.oldVersion.String() != foundVersion.String() {
				gotrace.Trace("%s : %s => %s\n", name, plugin.oldVersion, foundVersion)
			} else {
				gotrace.Trace("%s : %s => No update\n", name, plugin.oldVersion)
			}

		}
	}

	return true
}
