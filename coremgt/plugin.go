package coremgt

import (
	"fmt"
)

const (
	pluginType = "plugin"
)

// Plugin describe details on Plugin element.
// By default, it helps reading the MANIFEST.MF from java as a yaml file, cleaned before.
type Plugin struct {
	Version        string `yaml:"Plugin-Version"`
	ExtensionName  string `yaml:"Extension-Name"`
	ShortName      string `yaml:"Short-Name"`
	JenkinsVersion string `yaml:"Jenkins-Version"`
	LongName       string `yaml:"Long-Name"`
	Dependencies   string `yaml:"Plugin-Dependencies"`
	Description    string `yaml:"Specification-Title"`
}

// GetVersion return the plugin Version struct.
func (p *Plugin) GetVersion() (ret VersionStruct, err error) {
	err = ret.Set(p.Version)
	return
}

// NewPlugin return a plugin object
func NewPlugin() (ret *Plugin) {
	ret = new(Plugin)
	return
}

// SetFrom set data from an array of fields
func (p *Plugin) SetFrom(fields ...string) (err error) {
	fieldsSize := len(fields)
	if fieldsSize < 2 {
		err = fmt.Errorf("Invalid data type. Requires type (field 1) as '%s' and plugin name (field 2)", pluginType)
		return
	}
	if fields[0] != pluginType {
		err = fmt.Errorf("Invalid data type. Must be '%s'", pluginType)
		return
	}
	p.ExtensionName = fields[1]
	if fieldsSize >= 3 {
		p.Version = fields[2]
	}
	return
}

// GetType return the internal type string
func (p *Plugin) GetType() string {
	return pluginType
}

// Name return the Name property
func (p *Plugin) Name() string {
	return p.ExtensionName
}
