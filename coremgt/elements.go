package coremgt

import (
	"fmt"
	"net/url"

	"jplugins/simplefile"
	"sort"
	"strings"

	"github.com/forj-oss/utils"
	"github.com/forj-oss/forjj-modules/trace"
)

const (
	oldestPolicy = 0
	keepPolicy   = 1
	newestPolicy = 2
)

// Elements represents the list of Elements (plugins, features, etc...)
type Elements map[string]Element

// ElementsType is collections of Plugins
type ElementsType struct {
	list      map[string]Elements
	supported []string

	repoPath string
	repoURL  []*url.URL
	useLocal bool
	noDeps   bool

	ref *Repository
}

// NewElementsType creates the collection of plugins
func NewElementsType() (ret *ElementsType) {
	ret = new(ElementsType)
	ret.list = make(map[string]Elements)
	ret.supported = []string{featureType, groovyType, pluginType}
	ret.noDeps = false
	return
}

// SetRepository set a Repository object to the Elementstype object.
func (e *ElementsType) SetRepository(ref *Repository) {
	if e == nil {
		return
	}
	e.ref = ref
}

// noChainLoaded
func (e *ElementsType) noChainLoaded() {
	if e == nil {
		return
	}
	e.noDeps = true
}

// GetElements return the collection type requested.
func (e *ElementsType) GetElements(elementType string) (_ Elements) {
	if e == nil {
		return
	}
	if ret, found := e.list[elementType]; found {
		return ret
	}
	return
}

// AddElement a new element to a collection type.
func (e *ElementsType) AddElement(element Element) (ret Element, err error) {
	if e == nil {
		return
	}

	if element == nil {
		return
	}

	elementType := element.GetType()

	elements, found := e.list[elementType]

	if !found {
		elements = make(map[string]Element)
	}

	if existing, found := elements[element.Name()]; found {
		// Keep the most recent version (base constraint)
		element.Merge(existing, newestPolicy)
	}

	elements[element.Name()] = element
	e.list[elementType] = elements
	ret = element

	if e.noDeps {
		return
	}
	return ret, e.addChainedElements(element)
}

// Add a new element to a collection type.
func (e *ElementsType) Add(fields ...string) (_ Element, err error) {
	if e == nil {
		return
	}

	if len(fields) < 2 {
		err = fmt.Errorf("Invalid format. Requires at least 2 fields, ie type and name")
		return
	}
	elementType := fields[0]
	if !e.checkElementType(elementType) {
		err = fmt.Errorf("Unsupported elementType")
		return
	}

	return e.add(fields...)
}

// add internally the fields as a new element even if the root list restrict in types. (no call to checkElementType)
func (e *ElementsType) add(fields ...string) (element Element, err error) {
	elementType := fields[0]
	name := fields[1]

	elements, found := e.list[elementType]
	if !found {
		elements = make(map[string]Element)
		element = NewElement(elementType)
	} else if element, found = elements[name]; !found {
		element = NewElement(elementType)
	}
	err = element.SetFrom(fields...)
	if err != nil {
		return
	}

	element.CompleteFromContext(e)

	if existing, found := elements[element.Name()]; found {
		// Keep the most recent version (base constraint)
		element.Merge(existing, newestPolicy)
		gotrace.Trace("Updated %s", element)
	} else {
		gotrace.Trace("Added %s", element)
	}

	elements[name] = element
	e.list[elementType] = elements


	if e.noDeps {
		return
	}

	return element, e.addChainedElements(element)
}

// Remove a named element type.
func (e *ElementsType) Remove(elementType, name string) {
	if plugins, found := e.list[elementType]; found {
		if _, found := plugins[name]; found {
			delete(plugins, name)
			e.list[elementType] = plugins
		}
	}

}

// SetLocal set the useLocal to true
// When set to true, jplugin do not clone a remote repo URL to store on cache
func (e *ElementsType) SetLocal() {
	if e == nil {
		return
	}
	e.useLocal = true
}

// SetFeaturesRepoURL validate and store the repoURL
// It can be executed several times to store multiple URLs
func (e *ElementsType) SetFeaturesRepoURL(repoURL string) error {
	if e == nil {
		return nil
	}

	if repoURLObject, err := url.Parse(repoURL); err != nil {
		return fmt.Errorf("Invalid feature repository URL. %s", repoURL)
	} else {
		e.repoURL = append(e.repoURL, repoURLObject)
	}
	return nil
}

// SetFeaturesPath defines where a features repository is located like `jenkins-install-inits`
func (e *ElementsType) SetFeaturesPath(repoPath string) error {
	if e == nil {
		return nil
	}

	if p, err := utils.Abs(repoPath); err != nil {
		return fmt.Errorf("Invalid feature repository path. %s", repoPath)
	} else {
		e.repoPath = p
	}
	return nil
}

// AddSupport register a set function for a data type given
func (e *ElementsType) AddSupport(elementTypes ...string) {
	e.supported = elementTypes
}

// Read the file given
func (e *ElementsType) Read(file string, cols int) (err error) {
	data := simplefile.NewSimpleFile(file, cols)

	err = data.Read(":", func(fields []string) (err error) {
		_, err = e.Add(fields...)
		return
	})
	return
}

// ExtractTopElements identifies top plugins (remove all dependencies)
func (e *ElementsType) ExtractTopElements() (identified *ElementsType) {
	identified = NewElementsType()

	plugins := e.list[pluginType]
	for _, plugin := range plugins {
		identified.AddElement(plugin)
	}

	for _, plugin := range plugins {
		if plugin.(*Plugin).Dependencies == "" {
			continue
		}
		for _, depPluginDetail := range strings.Split(plugin.(*Plugin).Dependencies, ",") {
			depPlugin := strings.Split(depPluginDetail, ":")
			identified.Remove(pluginType, depPlugin[0])
		}
	}
	return

}

// WriteSimple the list of plugins as Simple format
func (e *ElementsType) WriteSimple(file string, cols int) (err error) {
	featureFile := simplefile.NewSimpleFile(file, 2)
	for elementType, plugins := range e.list {
		for name := range plugins {
			featureFile.Add(1, elementType, name)
		}
	}
	err = featureFile.WriteSorted(":")
	return
}

// Length the list of plugins as Simple format
func (e *ElementsType) Length() (total int) {
	for _, elements := range e.list {
		total += len(elements)
	}
	return
}

// PrintOut loop on plugins to display them
func (e *ElementsType) PrintOut(printDetails func(element Element)) {

	plugins := e.list[pluginType]
	pluginsList := make([]string, len(plugins))

	iCount := 0
	for name := range plugins {
		pluginsList[iCount] = name
		iCount++
	}

	sort.Strings(pluginsList)

	for _, name := range pluginsList {
		printDetails(plugins[name])
	}
}

/************************************************************************
 ***************** INTERNAL FUNCTIONS ***********************************
 ************************************************************************/

func (e *ElementsType) checkElementType(elementType string) (found bool) {
	for _, value := range e.supported {
		if value == elementType {
			found = true
			return
		}
	}
	return
}

func (e *ElementsType) addChainedElements(element Element) (_ error) {
	elementsType, err := element.ChainElement(e)
	if err != nil {
		err = fmt.Errorf("Unable to attach elements related to %s-%s. %s", element.GetType(), element.Name(), err)
		return
	}
	for _, elements := range elementsType.list {
		for _, element := range elements {
			if _, err = e.AddElement(element); err != nil {
				return
			}
		}
	}

	return
}
