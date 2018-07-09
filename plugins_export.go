package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/forj-oss/forjj-modules/trace"
)

type pluginsExport struct {
	json         []pluginJson
	exportFile   string
	templateFile string
	plugins      *pluginsStatus
}

type pluginsJson []pluginJson

type pluginJson struct {
	Name        string
	ShortName   string
	Description string
	OldVersion  string
	NewVersion  string
}

func newPluginsExport(exportFile, templateFile string, size int) (ret *pluginsExport) {
	ret = new(pluginsExport)
	ret.exportFile = exportFile
	ret.templateFile = templateFile
	ret.json = make([]pluginJson, 0, size)
	return
}

func (e *pluginsExport) doItOn(list *pluginsStatus) {
	e.plugins = list
	if e.templateFile == "" {
		e.doExportJSON()
	} else {
		e.doExportTemplate()
	}
	fmt.Printf("\nFound %d plugin(s) updates.\n", len(e.json))
}

func (e *pluginsExport) doExportJSON() {
	e.plugins.pluginsStatus = make(map[string]*pluginsStatusDetails)

	pluginsList, _ := e.plugins.sortPlugins()

	for _, plugin := range pluginsList {
		pluginInfo, found := e.plugins.pluginsStatus[plugin]
		if !found {
			continue
		}
		plugin := pluginJson{
			Name:        pluginInfo.name,
			Description: pluginInfo.title,
			OldVersion:  pluginInfo.oldVersion.String(),
			NewVersion:  pluginInfo.newVersion.String(),
		}
		e.json = append(e.json, plugin)
	}

	if jsonData, err := json.MarshalIndent(e.json, "", "  "); err != nil {
		gotrace.Error("Unable to encode in JSON. %s", err)
		return
	} else {
		err = ioutil.WriteFile(e.exportFile, jsonData, 0644)
		if err != nil {
			gotrace.Error("Unable to save %s. %s", e.exportFile, err)
			return
		}
		gotrace.Info("'%s' exported and saved.", e.exportFile)
	}
}

func (e *pluginsExport) doExportTemplate() {

}
