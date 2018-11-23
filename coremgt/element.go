package coremgt

// Element is an interface on Plugin/Groovy/Feature object
type Element interface{
	GetVersion() (ret VersionStruct, err error)
	SetFrom(fields ...string) (err error)
	GetType() string
	Name() string
}

// NewElement to create a known new element type
func NewElement(elementType string) (_ Element) {
	switch elementType {
	case "plugin":
		return NewPlugin()
	case "feature":
		return NewFeature()
	case "groovy":
		return NewGroovy()
	}
	return
}