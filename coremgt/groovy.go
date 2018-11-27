package coremgt

import (
	"fmt"
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
}

// NewGroovy return a Groovy object
func NewGroovy() (ret *Groovy) {
	ret = new(Groovy)
	return
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
