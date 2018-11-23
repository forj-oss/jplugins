package coremgt

import (
	"fmt"
)

const (
	featureType = "feature"
)

// Feature describe details on Feature.
type Feature struct {
	Version string
	name    string
}

// NewFeature return a feature object
func NewFeature() (ret *Feature) {
	ret = new(Feature)
	return
}

// GetVersion return the plugin Version struct.
func (p *Feature) GetVersion() (_ VersionStruct, _ error) {
	return
}

// SetFrom set data from an array of fields
func (p *Feature) SetFrom(fields ...string) (err error) {
	fieldsSize := len(fields)
	if fieldsSize < 2 {
		err = fmt.Errorf("Invalid data type. Requires type (field 1) as '%s' and feature name (field 2)", featureType)
		return
	}
	if fields[0] != featureType {
		err = fmt.Errorf("Invalid data type. Must be '%s'", featureType)
		return
	}
	p.name = fields[1]
	return
}

// GetType return the internal type string
func (p *Feature) GetType() string {
	return featureType
}

// Name return the Name property
func (p *Feature) Name() string {
	return p.name
}