package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/forj-oss/forjj-modules/trace"
)

type pluginsStatus struct {
	plugins map[string]pluginsStatusDetails
}

type pluginsStatusDetails struct {
	name       string
	title      string
	oldVersion string
	newVersion string
	rules      []string
}

func newPluginsStatus() (pluginsCompared *pluginsStatus) {
	pluginsCompared = new(pluginsStatus)
	pluginsCompared.plugins = make(map[string]pluginsStatusDetails)
	return
}

func (s *pluginsStatus) compare(installed plugins, ref *repository) {
	if ref == nil {
		return
	}

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

func (s *pluginsStatus) add(version string, plugin repositoryPlugin) {
	s.plugins[plugin.Name] = pluginsStatusDetails{
		name:       plugin.Name,
		title:      plugin.Title,
		oldVersion: version,
		newVersion: plugin.Version,
	}
}

func (s *pluginsStatus) obsolete(plugin *pluginManifest) {
	s.plugins[plugin.Name] = pluginsStatusDetails{
		name:       plugin.Name,
		title:      plugin.LongName,
		oldVersion: plugin.Version,
		newVersion: "obsolete",
	}

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
	pluginsDetails := make(map[string]pluginsStatusDetails)
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

	for _, title := range pluginsList {
		plugin := pluginsDetails[title]
		fmt.Printf("%-"+strconv.Itoa(iMaxTitle + 3)+"s : %-10s => %s\n", title+" ("+plugin.name+")", plugin.oldVersion, plugin.newVersion)
	}

	fmt.Printf("\nFound %d plugin(s) updates available.\n", iCount)

	return true
}
