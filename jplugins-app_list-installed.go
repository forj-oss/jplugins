package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
	"gopkg.in/yaml.v2"
)

func (a *jPluginsApp) doListInstalled() {
	if !a.readFromJenkins(*a.listInstalled.jenkinsHomePath) {
		return
	}
	if *a.listInstalled.preInstalled {
		a.saveVersionAsPreInstalled(*a.listInstalled.jenkinsHomePath, a.installedPlugins)
		return
	}
	a.printOutVersion(a.installedPlugins)
}

// readFromJenkins read manifest of each plugins and store information in a.installedPlugins
func (a *jPluginsApp) readFromJenkins(jenkinsHomePath string) (_ bool) {
	pluginsPath := path.Join(jenkinsHomePath, "plugins")

	a.installedPlugins = make(plugins)

	fEntries, err := ioutil.ReadDir(pluginsPath)

	if err != nil {
		gotrace.Error("Invalid Jenkins home '%s'. %s", pluginsPath, err)
		return
	}

	var fileRE, manifestRE *regexp.Regexp
	fileREDefine := `^(.*)\.[jh]pi*$`
	manifestREDefine := `([\w-]*: )(.*)\n`

	if re, err := regexp.Compile(fileREDefine); err != nil {
		gotrace.Error("Internal error. Regex '%s': %s", fileREDefine, err)
		return
	} else {
		fileRE = re

	}
	if re, err := regexp.Compile(manifestREDefine); err != nil {
		gotrace.Error("Internal error. Regex '%s': %s", fileREDefine, err)
		return
	} else {
		manifestRE = re
	}

	for _, fEntry := range fEntries {
		if fEntry.IsDir() {
			continue
		}

		if fileMatch := fileRE.FindAllStringSubmatch(fEntry.Name(), -1); fileMatch != nil {
			pluginFileName := fileMatch[0][0]
			pluginName := fileMatch[0][1]

			if pluginFileName != "" && pluginName == "" {
				gotrace.Error("Invalid file '%s'. Ignored.", pluginFileName)
				continue
			}

			pluginMetafile := path.Join(pluginsPath, pluginName, "META-INF", "MANIFEST.MF")

			if _, err := os.Stat(pluginMetafile); err != nil && os.IsNotExist(err) {
				// TODO: Ignored for now. but may need to extract the plugin file to get the version
				gotrace.Warning("Plugin '%s' found but not expanded. Ignored. (fix in next jplugins version)", pluginMetafile)
				continue
			}

			var manifest *pluginManifest

			if d, err := ioutil.ReadFile(pluginMetafile); err != nil {
				gotrace.Error("Unable to read file '%s'. %s. Ignored", pluginMetafile, err)
				continue
			} else {
				// Remove DOS format if exist
				data := strings.Replace(string(d), "\r", "", -1)
				// and remove new lines ('\n ')
				data = strings.Replace(data, "\n ", "", -1)
				// and escape "
				data = strings.Replace(data, "\"", "\\\"", -1)
				// and embrace value part with "
				data = manifestRE.ReplaceAllString(data, `$1"$2"`+"\n")

				manifest = new(pluginManifest)
				if err := yaml.Unmarshal([]byte(data), &manifest); err != nil {
					gotrace.Error("Unable to read file '%s' as yaml. %s. Ignored", pluginMetafile, err)
					fmt.Print(data)
					continue
				}
			}
			a.installedPlugins[manifest.Name] = manifest
		}
	}
	return true
}

func (a *jPluginsApp) printOutVersion(plugins plugins) (_ bool) {
	if a.installedPlugins == nil {
		return
	}

	pluginsList := make([]string, len(plugins))

	iCount := 0
	for name := range plugins {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	for _, name := range pluginsList {
		fmt.Printf("%s: %s\n", name, plugins[name].Version)
	}
	fmt.Println(iCount, "plugin(s)")
	return true
}

func (a *jPluginsApp) saveVersionAsPreInstalled(jenkinsHomePath string, plugins plugins) (_ bool) {
	if a.installedPlugins == nil {
		return
	}

	pluginsList := make([]string, len(plugins))

	iCount := 0
	for name := range plugins {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	preInstalledFile := path.Join(jenkinsHomePath, "jplugins-preinstalled.lst")
	piDescriptor, err := os.OpenFile(preInstalledFile, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		gotrace.Error("Unable to create '%s'. %s", preInstalledFile, err)
		return
	}

	defer piDescriptor.Close()

	for _, name := range pluginsList {
		fmt.Fprintf(piDescriptor, "plugin:%s:%s\n", name, plugins[name].Version)
	}
	fmt.Printf("%d plugin(s) saved in '%s'\n", iCount, preInstalledFile)
	return true
}
