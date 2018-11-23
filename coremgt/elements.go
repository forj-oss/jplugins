package coremgt

import (
	"fmt"
	"net/url"

	"jplugins/simplefile"
	"strings"
	"sort"

	"github.com/forj-oss/utils"
)

// Elements is collections of Plugins
type Elements struct {
	list      map[string]map[string]Element
	supported []string

	repoPath string
	repoURL  []*url.URL
	useLocal bool
}

// NewElements creates the collection of plugins
func NewElements() (ret *Elements) {
	ret = new(Elements)
	ret.list = make(map[string]map[string]Element)
	ret.supported = []string{featureType, groovyType, pluginType}
	return
}

// GetElements return the collection type requested.
func (e *Elements) GetElements(elementType string) (_ map[string]Element) {
	if e == nil {
		return
	}
	if ret, found := e.list[elementType]; found {
		return ret
	}
	return
}

// AddElement a new element to a collection type.
func (e *Elements) AddElement(element Element) (err error) {
	if e == nil {
		return
	}

	if element == nil {
		return
	}

	elementType := element.GetType()
	found := false
	for _, value := range e.supported {
		if value == elementType {
			found = true
			break
		}
	}
	if !e.checkElementType(elementType) {
		return fmt.Errorf("Unsupported elementType")
	}

	elements, found := e.list[elementType]

	if !found {
		elements = make(map[string]Element)
	}
	elements[element.Name()] = element
	e.list[elementType] = elements
	return
}

// Add a new element to a collection type.
func (e *Elements) Add(fields ...string) (err error) {
	if e == nil {
		return
	}

	if len(fields) < 2 {
		return fmt.Errorf("Invalid format. Requires at least 2 fields, ie type and name")
	}
	elementType := fields[0]
	if !e.checkElementType(elementType) {
		return fmt.Errorf("Unsupported elementType")
	}
	name := fields[1]

	elements, found := e.list[elementType]
	var element Element
	if !found {
		elements = make(map[string]Element)
		element = NewElement(elementType)
	} else if element, found = elements[name]; !found {
		element = NewElement(elementType)
	}
	elements[name] = element
	e.list[elementType] = elements
	err = element.SetFrom(fields...)
	if err != nil {
		return
	}
	return
}

func (e *Elements) Remove(elementType, name string) {
	if plugins, found := e.list[elementType]; found {
		if _, found := plugins[name]; found {
			delete(plugins, name)
			e.list[elementType] = plugins
		}
	}

}

// SetLocal set the useLocal to true
// When set to true, jplugin do not clone a remote repo URL to store on cache
func (e *Elements) SetLocal() {
	if e == nil {
		return
	}
	e.useLocal = true
}

// SetFeaturesRepoURL validate and store the repoURL
// It can be executed several times to store multiple URLs
func (e *Elements) SetFeaturesRepoURL(repoURL string) error {
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
func (e *Elements) SetFeaturesPath(repoPath string) error {
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
func (e *Elements) AddSupport(elementTypes ...string) {
	e.supported = elementTypes
}

// Read the file given
func (e *Elements) Read(file string, cols int) (err error) {
	data := simplefile.NewSimpleFile(file, cols)

	err = data.Read(":", func(fields []string) error {
		return e.Add(fields...)
	})
	return
}

// ExtractTopElements identifies top plugins (remove all dependencies)
func (e *Elements) ExtractTopElements() (identified *Elements) {
	identified = NewElements()

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
func (e *Elements) WriteSimple(file string, cols int) (err error) {
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
func (e *Elements) Length() (total int) {
	for _, elements := range e.list {
		total += len(elements)
	}
	return
}

// PrintOut loop on plugins to display them
func (e *Elements) PrintOut(printDetails func(element Element)) {

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

func (e *Elements) checkElementType(elementType string) (found bool) {
	for _, value := range e.supported {
		if value == elementType {
			found = true
			return
		}
	}
	return
}
