package coremgt

import (
	"fmt"
	"regexp"
	"strings"

	goversion "github.com/hashicorp/go-version"

	"github.com/forj-oss/forjj-modules/trace"
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
	rules          map[string]goversion.Constraints
	fixed          bool // true if a constraint force a version
}

// String return the string representation of the plugin
func (p *Plugin) String() string {
	if p == nil {
		return "nil"
	}
	ruleShown := make([]string, len(p.rules))
	index := 0
	for _, rule := range p.rules {
		ruleShown[index] = rule.String()
		index++
	}

	return fmt.Sprintf("%s:%s %s (constraints: %s)", pluginType, p.ExtensionName, p.Version, strings.Join(ruleShown, ", "))
}

// GetVersion return the plugin Version struct.
func (p *Plugin) GetVersion() (ret VersionStruct, err error) {
	err = ret.Set(p.Version)
	return
}

// GetVersionString return the plugin value as string.
func (p *Plugin) GetVersionString() string {
	return p.Version
}

// NewPlugin return a plugin object
func NewPlugin() (ret *Plugin) {
	ret = new(Plugin)
	ret.rules = make(map[string]goversion.Constraints)
	return
}

// SetFrom set data from an array of fields
// If the version is given, it will be interpreted as a constraint
func (p *Plugin) SetFrom(fields ...string) (err error) {
	err = p.setFrom(fields...)
	if err != nil {
		return
	}

	if p.Version != "" { // If version is given, it will be an equal constraint, except if it is already a constraint
		var constraints goversion.Constraints
		constraints, err = goversion.NewConstraint(p.Version)
		if err != nil {
			err = fmt.Errorf("Version constraints are invalid. %s", err)
			return
		}

		p.rules[constraints.String()] = constraints

		constraintPiecesRe, _ := regexp.Compile(`^([<>=!~])*(.*)$`)
		constraintPieces := constraintPiecesRe.FindStringSubmatch(p.Version)
		if constraintPieces != nil {
			if constraintPieces[1] != "" {
				// Remove the constraints rule piece of the verison string
				p.Version = constraintPieces[2]
			}
			if constraintPieces[0] == "" || constraintPieces[0] == "=" {
				p.fixed = true
			}
		}
	}
	return
}

// setFrom set data from an array of fields, with no constraints
func (p *Plugin) setFrom(fields ...string) (err error) {
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

// CompleteFromContext update the plugin information from repo DB if found
func (p *Plugin) CompleteFromContext(context *ElementsType) {
	if p == nil || context == nil || context.ref == nil {
		return
	}

	refPlugin, found := context.ref.Get(p.ExtensionName, p.Version)
	if !found {
		return
	}

	p.ShortName = refPlugin.Name
	p.JenkinsVersion = refPlugin.JenkinsVersion
	p.Description = refPlugin.Description

}

// GetType return the internal type string
func (p *Plugin) GetType() string {
	return pluginType
}

// Name return the Name property
func (p *Plugin) Name() string {
	return p.ExtensionName
}

// ChainElement load plugins dependency tree from the repo
//
func (p *Plugin) ChainElement(context *ElementsType) (ret *ElementsType, _ error) {
	gotrace.Trace("building chained list from %s", p)
	version := "latest"
	if p.Version != "latest" {
		version = p.Version
	}
	refPlugin, found := context.ref.Get(p.ExtensionName, version)
	if !found {
		return nil, fmt.Errorf("Plugin '%s' not found in the public repository", p.Name())
	}

	ret = NewElementsType()
	ret.AddSupport(pluginType)
	ret.noChainLoaded()
	ret.SetRepository(context.ref)

	for _, dep := range refPlugin.Dependencies {
		if dep.Optionnal {
			continue
		}
		plugin := NewPlugin()
		plugin.SetFrom(pluginType, dep.Name, ">="+dep.Version)
		ret.AddElement(plugin)
	}
	gotrace.Trace("Chained list built from %s", p)
	return
}

// Merge execute a merge between 2 plugins and keep the one corresponding to the constraint given
// It is based on 3 policies: choose oldest, keep existing and choose newest
func (p *Plugin) Merge(element Element, policy int) (updated bool, err error) {
	if p == nil {
		return
	}
	if p.fixed { // The plugin version is fixed (= constraint)
		return
	}

	switch policy {
	case oldestPolicy:
		origVersion, _ := p.GetVersion()
		newPlugin, ok := element.(*Plugin)
		if !ok {
			err = fmt.Errorf("plugin merge support only plugins element type.")
			return
		}
		newVersion, _ := newPlugin.GetVersion()

		if origVersion.Get().GreaterThan(newVersion.Get()) {
			p.Version = newPlugin.Version
			p.rules = newPlugin.rules
			updated = true
		}
	case keepPolicy:
	case newestPolicy:
		origVersion, _ := p.GetVersion()
		newPlugin, ok := element.(*Plugin)
		if !ok {
			err = fmt.Errorf("plugin merge support only plugins element type.")
			return
		}
		newVersion, _ := newPlugin.GetVersion()

		if origVersion.Get().LessThan(newVersion.Get()) {
			p.Version = newPlugin.Version
			p.rules = newPlugin.rules
			updated = true
		}
	}
	return
}

// IsFixed indicates if the plugin version is fixed.
func (p *Plugin) IsFixed() (_ bool) {
	if p == nil {
		return
	}
	return p.fixed
}
