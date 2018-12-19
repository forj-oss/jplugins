package coremgt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/forj-oss/forjj-modules/trace"
)

type PluginsExport struct {
	json         pluginsJson
	exportFile   string
	templateFile string
	plugins      *PluginsStatus
}

func NewPluginsExport(exportFile, templateFile string, size int) (ret *PluginsExport) {
	ret = new(PluginsExport)
	ret.exportFile = exportFile
	ret.templateFile = templateFile
	ret.json = make([]pluginJson, 0, size)
	return
}

func (e *PluginsExport) DoItOn(list *PluginsStatus) (err error) {
	e.plugins = list
	e.buildList()

	if e.templateFile == "" {
		err = e.doExportJSON()
	} else {
		err = e.doExportTemplate()
	}
	if err != nil {
		return
	}
	fmt.Printf("\nFound %d plugin(s) updates.\n", len(e.json))
	return
}

func (e *PluginsExport) buildList() {
	e.plugins.PluginsStatus = make(map[string]*pluginsStatusDetails)

	pluginsList, _ := e.plugins.sortPlugins()

	for _, plugin := range pluginsList {
		pluginInfo, found := e.plugins.PluginsStatus[plugin]
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
}

func (e *PluginsExport) doExportJSON() error {
	jsonData, err := json.MarshalIndent(e.json, "", "  ")
	if err != nil {
		return fmt.Errorf("Unable to encode in JSON. %s", err)
	}
	err = ioutil.WriteFile(e.exportFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("Unable to save %s. %s", e.exportFile, err)

	}
	gotrace.Info("'%s' exported and saved.", e.exportFile)
	return nil
}

func (e *PluginsExport) doExportTemplate() error {
	tmplData, err := ioutil.ReadFile(e.templateFile)
	if err != nil {
		return fmt.Errorf("'%s' unreadable. %s", e.templateFile, err)
	}

	var exportFile *os.File
	exportFile, err = os.Create(e.exportFile)
	if err != nil {
		return fmt.Errorf("'%s' unreadable. %s", e.templateFile, err)

	}
	defer exportFile.Close()

	tmpl := template.New("export template").Funcs(template.FuncMap{
		"isLast": func(index, length int) bool {
			return (index+1 >= length)
		},
	})

	_, err = tmpl.Parse(strings.Replace(string(tmplData), "\\\n", "", -1))
	if err != nil {
		return fmt.Errorf("template '%s' has errors. %s", e.templateFile, err)

	}

	err = tmpl.Execute(exportFile, &e.json)
	if err != nil {
		return fmt.Errorf("template execution '%s' has errors. %s", e.templateFile, err)

	}

	gotrace.Info("'%s' exported and saved from template '%s'.", e.exportFile, e.templateFile)
	return nil
}
