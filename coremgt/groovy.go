package coremgt

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	git "github.com/forj-oss/go-git"

	"github.com/forj-oss/forjj-modules/trace"
	goversion "github.com/hashicorp/go-version"
)

const (
	groovyType = "groovy"
)

// Groovy describe details on Groovy.
type Groovy struct {
	Version       string
	name          string
	ShortName     string
	Dependencies  string
	Description   string
	CommitID      string
	Md5           string
	rules         map[string]goversion.Constraints
	commitHistory []string

	featureName string
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

	return fmt.Sprintf("%s:%s %s (constraints: %s)", groovyType, p.name, p.Version, strings.Join(ruleShown, ", "))
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
func (p *Groovy) CompleteFromContext(context *ElementsType) (err error) {
	if context == nil {
		return
	}
	if v, found := context.supportContext[groovyType]; found {
		if p.featureName, found = v["featureName"]; !found {
			gotrace.Error("Invalid groovy code reference. It must be defined by feature. supportContext['featureName'] not defined.")
			return
		}
	} else {
		gotrace.Error("Invalid groovy code reference. It must be defined by feature. supportContext is nil or not defined for 'groovy', thus missing 'featureName'.")
		return
	}

	sourcePath := path.Join(context.repoPath, p.featureName)

	if err = p.defineVersion(sourcePath) ; err != nil {
		return
	}
	err = p.computeM5Sum(sourcePath)
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

// Merge execute a merge between 2 groovies and keep the one corresponding to the constraint given
// It is based on 3 policies: choose oldest, keep existing and choose newest
func (p *Groovy) Merge(_ *ElementsType, _ Element, _ int) (_ bool, _ error) {

	return
}

// IsFixed indicates if the groovy version is fixed.
func (p *Groovy) IsFixed() (_ bool) {
	return
}

// GetParents return the list of features which depends on this feature.
func (p *Groovy) GetParents() (_ Elements) {
	return
}

// GetDependencies return the list of features depedencies required by this feature.
func (p *Groovy) GetDependencies() (_ Elements) {
	return
}

// GetDependenciesFromContext return the list of features depedencies required by this feature.
func (p *Groovy) GetDependenciesFromContext(*ElementsType) (_ Elements) {
	return
}

// SetVersionConstraintFromDepConstraint add a constraint to match the
// dependency version constraints
func (p *Groovy) SetVersionConstraintFromDepConstraint(*ElementsType, Element) (_ error) {
	return
}

// IsVersionCandidate return true systematically
func (p *Groovy) IsVersionCandidate(version *goversion.Version) bool {
	return true
}

func (p *Groovy) RemoveDependencyTo(depElement Element) {
}

func (p *Groovy) AddDependencyTo(depElement Element) {
}

func (p *Groovy) DefineLatestPossibleVersion(context *ElementsType) (_ error) {
	return
}

// AsNewGrooviesStatusDetails add the current groovy as a NEW groovy in statusDetails
func (p *Groovy) AsNewGrooviesStatusDetails(context *ElementsType) (sd *GroovyStatusDetails) {
	sd = newGroovyStatusDetails(p.name, context.repoPath)
	sd.name = p.name
	sd.newMd5 = p.Md5
	sd.newCommit = p.CommitID
	return
}

// AsNewPluginsStatusDetails to be replaced by a renamed version.
func (p *Groovy) AsNewPluginsStatusDetails(context *ElementsType) (sd *pluginsStatusDetails) {
	return
}

// computeM5Sum get the groovy file md5sum
func (p *Groovy) computeM5Sum(sourcePath string) (_ error) {
	groovyFile := path.Join(sourcePath, p.name+".groovy")
	fd, err := os.Open(groovyFile)
	if err != nil {
		return fmt.Errorf("Unable to read '%s'. %s", groovyFile, err)
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	hash := md5.New()

	if _, err := io.Copy(hash, reader); err != nil {
		return fmt.Errorf("Unable to generate md5sum data. %s", err)
	}
	md5Data := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	p.Md5 = md5Data

	return
}

// defineVersion get the latest commit ID updating the groovy file
func (p *Groovy) defineVersion(sourcePath string) (err error) {
	if p.commitHistory == nil {
		err = git.RunInPath(sourcePath, func() (_ error) {
			historyData, err := git.Get("log", "--pretty=%H", p.name+".groovy")
			if err != nil {
				return fmt.Errorf("Unable to get file '%s' history from GIT. %s", p.name+".groovy", err)
			}
			p.commitHistory = strings.Split(strings.Trim(historyData, " \n"), "\n")
			return
		})
		if err != nil {
			err = fmt.Errorf("Unable to define the groovy '%s' version (commit ID)> %s", p.name, err)
			return
		}
	}
	if len(p.commitHistory) == 0 {
		return
	}
	latest := p.commitHistory[0]
	p.CommitID = latest
	return
}
