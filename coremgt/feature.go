package coremgt

import (
	"fmt"
	"jplugins/simplefile"
	"path"

	"github.com/forj-oss/go-git"
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

// ChainElement load the feature details (groovy and plugins)
func (p *Feature) ChainElement(context *ElementsType) (ret *ElementsType, _ error) {
	if p == nil || context == nil {
		return
	}

	ret = NewElementsType()
	ret.AddSupport(pluginType, groovyType)
	ret.noChainLoaded()

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

	err := simpleFile.Read(":", func(fields []string) error {
		return ret.Add(fields...)
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to read feature file '%s'. %s", featureFile, err)
	}
	return
}
