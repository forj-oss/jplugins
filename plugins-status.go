package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	goversion "github.com/hashicorp/go-version"

	"github.com/forj-oss/forjj-modules/trace"
)

type pluginsStatus struct {
	plugins   map[string]*pluginsStatusDetails
	installed plugins
	ref       *repository
}

func newPluginsStatus(installed plugins, ref *repository) (pluginsCompared *pluginsStatus) {
	pluginsCompared = new(pluginsStatus)
	pluginsCompared.plugins = make(map[string]*pluginsStatusDetails)
	pluginsCompared.installed = installed
	pluginsCompared.ref = ref
	return
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

func (s *pluginsStatus) checkElement(line string) {
	var ftype, fname string
	var fversion string

	fields := strings.Split(line, ":")
	switch len(fields) {
	case 1:
		gotrace.Warning("Line %d: Line format is incorrect. It should be <'plugin'|'feature'>:<plugin Name>[:<version>]")
		return
	case 2:
		ftype = strings.Trim(fields[0], " ")
		fname = strings.Trim(fields[1], " ")
	default:
		ftype = strings.Trim(fields[0], " ")
		fname = strings.Trim(fields[1], " ")
		fversion = strings.Trim(fields[2], " ")
	}

	if ftype != "plugin" {
		gotrace.Warning("feature type '%s' is currently not supported. Ignored.", ftype)
		return
	}

	s.checkPlugin(fname, fversion)
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
