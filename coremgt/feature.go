package coremgt

import (
	"fmt"
	"jplugins/simplefile"
	"path"
	"strings"

	"github.com/forj-oss/go-git"
	goversion "github.com/hashicorp/go-version"
)

const (
	featureType = "feature"
)

// Feature describe details on Feature.
type Feature struct {
	Version string
	name    string
	rules   map[string]goversion.Constraints
}

// NewFeature return a feature object
func NewFeature() (ret *Feature) {
	ret = new(Feature)
	return
}

// String return the string representation of the plugin
func (p *Feature) String() string {
	if p == nil {
		return "nil"
	}
	ruleShown := make([]string, len(p.rules))
	index := 0
	for _, rule := range p.rules {
		ruleShown[index] = rule.String()
		index++
	}

	return fmt.Sprintf("%s:%s %s (constraints: %s)", featureType, p.name, p.Version, strings.Join(ruleShown, ", "))
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

// CompleteFromContext nothing to complete.
func (p *Feature) CompleteFromContext(_ *ElementsType) {
}

// GetType return the internal type string
func (p *Feature) GetType() string {
	return featureType
}

// Name return the Name property
func (p *Feature) Name() string {
	return p.name
}

// ChainElement load the feature details (groovy and plugins)
func (p *Feature) ChainElement(context *ElementsType) (ret *ElementsType, _ error) {
	if p == nil || context == nil {
		return
	}

	ret = NewElementsType()
	ret.AddSupport(pluginType, groovyType)
	ret.noRecursiveChain()
	ret.SetRepository(context.ref)

	if !context.useLocal {
		return nil, fmt.Errorf("Git clone of repository not currently implemented. Do git task and use --features-repo-path")
	}

	if err := git.RunInPath(context.repoPath, func() error {
		if git.Do("rev-parse", "--git-dir") != 0 {
			return fmt.Errorf("Not a valid GIT repository")
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("Issue with '%s', %s", context.repoPath, err)
	}

	featureFile := path.Join(context.repoPath, p.Name(), p.Name()+".desc")

	simpleFile := simplefile.NewSimpleFile(featureFile, 3)

	err := simpleFile.Read(":", func(fields []string) (err error) {
		_, err = ret.Add(fields...)
		return
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to read feature file '%s'. %s", featureFile, err)
	}
	return
}

// Merge execute a merge between 2 features and keep the one corresponding to the constraint given
// It is based on 3 policies: choose oldest, keep existing and choose newest
func (p *Feature) Merge(_ *ElementsType, _ Element, _ int) (_ bool, _ error) {

	return
}

// IsFixed indicates if the plugin version is fixed.
func (p *Feature) IsFixed() (_ bool) {
	return
}

// GetParents return the list of features which depends on this feature.
func (p *Feature) GetParents() (_ map[string]Element) {
	return
}

// GetDependencies return the list of features depedencies required by this feature.
func (p *Feature) GetDependencies() (_ map[string]Element) {
	return
}

// GetDependenciesFromContext return the list of features depedencies required by this feature.
func (p *Feature) GetDependenciesFromContext(*ElementsType) (_ map[string]Element) {
	return
}

// SetVersionConstraintFromDepConstraint add a constraint to match the
// dependency version constraints
func (p *Feature) SetVersionConstraintFromDepConstraint(*ElementsType, Element) (_ error) {
	return
}

// IsVersionCandidate return true systematically
func (p *Feature) IsVersionCandidate(version *goversion.Version) bool {
	return true
}

func (p *Feature) RemoveDependencyTo(depElement Element) {
}

func (p *Feature) AddDependencyTo(depElement Element) {
}