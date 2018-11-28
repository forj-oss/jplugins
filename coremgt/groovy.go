package coremgt

import (
	"fmt"
	"strings"
	goversion "github.com/hashicorp/go-version"
)

const (
	groovyType = "groovy"
)

// Groovy describe details on Groovy.
type Groovy struct {
	Version      string
	name         string
	ShortName    string
	Dependencies string
	Description  string
	CommitID     string
	rules          map[string]goversion.Constraints
}

// NewGroovy return a Groovy object
func NewGroovy() (ret *Groovy) {
	ret = new(Groovy)
	return
}

// String return the string representation of the plugin
func (p *Groovy) String() string {
        if p == nil {
                return "nil"
        }
	ruleShown := make([]string, len(p.rules))
        index := 0
        for _, rule := range p.rules {
                ruleShown[index] = rule.String()
                index++
        }

        return fmt.Sprintf("%s:%s-%s (constraints: %s)\n", pluginType, p.name, p.Version, strings.Join(ruleShown, ", "))
}

// GetVersion return the plugin Version struct.
func (p *Groovy) GetVersion() (_ VersionStruct, _ error) {
	return
}

// SetFrom set data from an array of fields
func (p *Groovy) SetFrom(fields ...string) (err error) {
	fieldsSize := len(fields)
	if fieldsSize < 2 {
		err = fmt.Errorf("Invalid data type. Requires type (field 1) as '%s' and groovy name (field 2)", groovyType)
		return
	}
	if fields[0] != groovyType {
		err = fmt.Errorf("Invalid data type. Must be '%s'", groovyType)
		return
	}
	p.name = fields[1]
	if fieldsSize >= 3 {
		p.CommitID = fields[2]
	}
	return
}

// CompleteFromContext nothing to complete.
func (p *Groovy)CompleteFromContext(_ *ElementsType) {
}

// GetType return the internal type string
func (p *Groovy) GetType() string {
	return groovyType
}

// Name return the Name property
func (p *Groovy) Name() string {
	return p.name
}

// ChainElement do nothing for a Groovy object
func (p *Groovy) ChainElement(*ElementsType) (_ *ElementsType, _ error) {
	return
}

// Merge execute a merge between 2 groovies and keep the one corresponding to the constraint given
// It is based on 3 policies: choose oldest, keep existing and choose newest
func (p *Groovy) Merge(element Element, policy int) (err error) {

	return
}
