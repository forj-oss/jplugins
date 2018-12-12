package coremgt

import (
	"jplugins/simplefile"
	"sort"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
)

// Plugins contains a collection of plugins manifest. (see ElementManifest)
type Plugins struct {
	list      map[string]*ElementManifest
	supported map[string]func(data []string) (string, *ElementManifest)
}

// NewPlugins creates a Plugins struct
func NewPlugins() (ret *Plugins) {
	ret = new(Plugins)
	ret.list = make(map[string]*ElementManifest)
	ret.supported = make(map[string]func(data []string) (string, *ElementManifest))
	return
}

// AddSupport register a set function for a data type given
func (p *Plugins) AddSupport(name string, set func(data []string) (string, *ElementManifest)) {
	p.supported[name] = set
}

// Read the file given
func (p *Plugins) Read(file string, cols int) (err error) {
	data := simplefile.NewSimpleFile(file, cols)

	err = data.Read(":", func(pluginRecord []string) (_ error) {
		setFunc, found := p.supported[pluginRecord[0]]
		if !found {
			return
		}
		name, plugin := setFunc(pluginRecord)
		p.list[name] = plugin
		return
	})
	return
}

// AddSupportPlugin define how plugin type data is extracted.
func (p *Plugins) AddSupportPlugin(repoGet func(name string) (*RepositoryPlugin, bool)) {
	p.AddSupport("plugin", func(pluginRecord []string) (index string, pluginData *ElementManifest) {
		index, pluginData = ExtractPlugin(pluginRecord)

		if refPlugin, found := repoGet(pluginData.Name); !found {
			gotrace.Warning("plugin '%s' is not recognized. Ignored.")
		} else {
			pluginData.LongName = refPlugin.Title
			pluginData.Description = refPlugin.Description
		}
		return
	})

}

// AddSupportGroovy define how groovy type data is extracted
func (p *Plugins) AddSupportGroovy() {
	p.AddSupport("groovy", ExtractGroovy)
}

// Length returns the number of plugins list
func (p *Plugins) Length() (_ int) {
	if p == nil {
		return
	}
	return len(p.list)
}

// ExtractTopElements identifies top plugins (remove all dependencies)
func (p *Plugins) ExtractTopElements() (identified *Plugins) {
	identified = NewPlugins()

	for name, plugin := range p.list {
		identified.list[name] = plugin
	}

	for _, plugin := range p.list {
		if plugin.Dependencies == "" {
			continue
		}
		for _, depPluginDetail := range strings.Split(plugin.Dependencies, ",") {
			depPlugin := strings.Split(depPluginDetail, ":")
			if _, found := identified.list[depPlugin[0]]; found {
				delete(identified.list, depPlugin[0])
			}
		}
	}
	return

}

// WriteSimple the list of plugins as Simple format
func (p *Plugins) WriteSimple(file string, cols int) (err error) {
	featureFile := simplefile.NewSimpleFile(file, 2)
	for name := range p.list {
		featureFile.AddWithKeyIndex(1, "plugin", name)
	}
	err = featureFile.WriteSorted(":")
	return
}

// PrintOut loop on plugins to display them
func (p *Plugins) PrintOut(printDetails func(element *ElementManifest)) {

	pluginsList := make([]string, len(p.list))

	iCount := 0
	for name := range p.list {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	for _, name := range pluginsList {
		printDetails(p.list[name])
	}
}

func (p *Plugins) Add(fields ...string) (_ error) {
	return
}
