package coremgt

import (
	goversion "github.com/hashicorp/go-version"
)

// Element is an interface on Plugin/Groovy/Feature object
type Element interface{
	GetVersion() (ret VersionStruct, err error)
	SetFrom(fields ...string) (err error)
	GetType() string
	Name() string
	ChainElement(context *ElementsType) (*ElementsType, error)
	//GetElementsFromRepo(func (name string) (*Elements, error))
	Merge(context *ElementsType, element Element, policy int) (updated bool, err error)
	CompleteFromContext(context *ElementsType)
	String() string
	IsFixed() bool
	GetParents() Elements

	GetDependencies() Elements
	GetDependenciesFromContext(context *ElementsType) Elements
	AddDependencyTo(element Element)
	RemoveDependencyTo(element Element)

	SetVersionConstraintFromDepConstraint(context *ElementsType, element Element) error
	IsVersionCandidate(version *goversion.Version) bool
	DefineLatestPossibleVersion(context *ElementsType) (_ error)

	AsNewPluginsStatusDetails(context *ElementsType) (sd *pluginsStatusDetails)
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