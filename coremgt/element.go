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
	GetParents() map[string]Element

	GetDependencies() map[string]Element
	GetDependenciesFromContext(context *ElementsType) map[string]Element
	AddDependencyTo(element Element)
	RemoveDependencyTo(element Element)

	SetVersionConstraintFromDepConstraint(context *ElementsType, element Element) error
	IsVersionCandidate(version *goversion.Version) bool
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