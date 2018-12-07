package coremgt

import (
	"fmt"
	"net/url"

	"jplugins/simplefile"
	"sort"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/utils"
	goversion "github.com/hashicorp/go-version"
)

const (
	oldestPolicy = 0
	keepPolicy   = 1
	newestPolicy = 2
)

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

// noRecursiveChainLoaded
func (e *ElementsType) noRecursiveChain() {
	if e == nil {
		return
	}
	e.noDeps = true
}

// ********** Elements management *******************

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

// GetElement return the collection type requested.
func (e *ElementsType) GetElement(elementType, name string) (_ Element) {
	if e == nil {
		return
	}
	elements := e.GetElements(elementType)
	if ret, found := elements[name]; found {
		return ret
	}
	return
}

// AddElement the new element to the collection, then add their dependencies if noDeps is false.
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

	elements[element.Name()] = element
	e.list[elementType] = elements
	ret = element

	if e.noDeps {
		return
	} else {
		gotrace.Trace("Added %s.", element)
	}
	return ret, e.addChainedElements(element)
}

// UpdateElement the new element to the collection, then add their dependencies if noDeps is false.
func (e *ElementsType) UpdateElement(curElement, newElement Element) (updated bool, err error) {
	if e == nil {
		return
	}

	if curElement == nil || newElement == nil {
		return
	}

	elementType := curElement.GetType()
	elementName := curElement.Name()

	elements, found := e.list[elementType]

	if !found {
		elements = make(map[string]Element)
	}

	if _, found := elements[curElement.Name()]; !found {
		return false, fmt.Errorf("Element %s not registered", curElement)
	}

	version1, _ := curElement.GetVersion()
	version2, _ := newElement.GetVersion()
	gotrace.Trace("Updating existing %s %s: %s <=> %s.", elementType, elementName, version1, version2)

	updated, err = curElement.Merge(e, newElement, newestPolicy)
	if err != nil {
		return false, err
	} else if updated {
		gotrace.Trace("Updated: %s.", curElement)
	} else {
		gotrace.Trace("kept   : %s.", curElement)
	}

	if e.noDeps {
		return
	}

	err = e.addChainedElements(curElement)
	return
}

// DeleteElement remove the element from the list and cleanup dependencies
func (e *ElementsType) DeleteElement(element Element) {
	elements := e.GetElements(element.GetType())
	gotrace.Trace("Deleting %s", element)
	delete(elements, element.Name())

	for _, depElement := range element.GetDependencies() {
		depElement.RemoveDependencyTo(element)
		if len(depElement.GetParents()) == 0 {
			e.DeleteElement(depElement)
		}
	}
}

// RefreshDependencies refresh Dependencies between old element list and new element list
func (e *ElementsType) RefreshDependencies(curElement, newElement Element) {
	// Remove old dependencies
	curDeps := curElement.GetDependencies()
	newDeps := newElement.GetDependenciesFromContext(e)
	if gotrace.IsDebugMode() {
		fmt.Printf("Current Deps:\n")
		curDeps.DisplayDependencies()
		fmt.Printf("New Deps:\n")
		newDeps.DisplayDependencies()
	}
	for name, curDependency := range curDeps {
		if _, found := newDeps[name]; !found {
			curElement.RemoveDependencyTo(curDependency)
		}
	}

	// Add Missing new dependencies
	for name, newDependency := range newDeps {
		if _, found := curDeps[name]; !found {
			curElement.AddDependencyTo(newDependency)
		}
	}
}

// ************** Element management by name ***************

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

// Remove a named element type.
func (e *ElementsType) Remove(elementType, name string) {
	if plugins, found := e.list[elementType]; found {
		if _, found := plugins[name]; found {
			delete(plugins, name)
			e.list[elementType] = plugins
		}
	}
}

// *********** Elementtype list configuration **************

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

// **************** Misc ************************************

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
	identified.noRecursiveChain()

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

// DeterminePluginsVersion apply the updates repository ref to the list of plugins to set proper version of plugins from rules
// Rules can change and a rule validation will be executed at the end of the process
func (e *ElementsType) DeterminePluginsVersion(ref *Repository) (updates *PluginsStatus, _ error) {
	if e == nil {
		return
	}
	if ref == nil {
		ref = e.ref
	}

	// Setting parent plugin constraints due to fixed version plugin
	e.defineParentConstraints()

/*	if gotrace.IsDebugMode() {
		fmt.Println(" ****** Identifying latest version from constraints given *******")
	}
	// Build latest versions from rules given
	queue := make(map[string]bool)

	for {
		round := len(queue)
		gotrace.Trace("%d/%d (%d) elements treated.", round, len(queue), e.Length())

		for _, elements := range e.list {
			for _, element := range elements {
				queueKey := element.GetType() + ":" + element.Name()
				if _, found := queue[queueKey]; found {
					continue
				}
				queue[queueKey] = true
				if err := element.DefineLatestPossibleVersion(e); err != nil {
					return nil, err
				}
				if err := e.addChainedElements(element); err != nil {
					return nil, err
				}
			}
		}
		if round == len(queue) {
			break
		}
	}*/

	updates = NewPluginsStatus(e, e.ref)
	// Consider element list as part of a new install.
	updates.NewInstall()

	return
}

// GetRepoPlugin return a plugin information from the updates repository
func (e *ElementsType) GetRepoPlugin(props ...string) (ret Element) {
	ret = NewPlugin()
	ret.SetFrom(append([]string{pluginType}, props...)...)
	ret.CompleteFromContext(e)
	return
}

/************************************************************************
 ***************** INTERNAL FUNCTIONS ***********************************
 ************************************************************************/

// defineParentConstraints is used to define the parent constraints
// required by a plugin version pinned.
// plugins which depends on a fixed version plugin may not support latest version, so those plugins must be downgraded.
func (e *ElementsType) defineParentConstraints() (_ error) {
	if gotrace.IsDebugMode() {
		fmt.Println("********** Updating parent constraints from fixed versions ************")
	}
	fixedElements := e.getFixedElements()

	for _, element := range fixedElements {
		e.setElementsRequiredVersion(element, element.GetParents())
	}
	return
}

// getFixedElements return a list of elements having a pinned version
func (e *ElementsType) getFixedElements() (fixedElements []Element) {

	fixedElements = make([]Element, 0, 8)

	for _, elements := range e.list {
		for _, element := range elements {
			if element.IsFixed() {
				fixedElements = append(fixedElements, element)
			}
		}
	}
	return
}

// setElementsRequiredVersion recursively will set an highest version following the dependency constraint given.
// This function is called by define ParentConstraints to browse to the higher dependent elements which may need to
// have constraint updated.
func (e *ElementsType) setElementsRequiredVersion(element Element, elements map[string]Element) {
	for _, elementToConstrain := range elements {
		elementToConstrain.SetVersionConstraintFromDepConstraint(e, element)
	}
}

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

	// Build a list of plugins required by the current one.
	// Each plugins listed has not been built recursively
	// To get sub deps on each dep, must be done with GetDependenciesFromContext(context)
	elementTypeDeps, err := element.ChainElement(e)

	if err != nil {
		return err
	}

	if gotrace.IsDebugMode() {
		if len(elementTypeDeps.list) == 0 {
			fmt.Printf("%s has no dependencies.\n", element)
		} else {
			fmt.Printf("%s has:\n", element)
			for _, elementDeps := range elementTypeDeps.list {
				for _, elementDependency := range elementDeps {
					elementDepType := elementDependency.GetType()
					elementDepName := elementDependency.Name()
					foundStr := ""
					if existingDep := e.GetElement(elementDepType, elementDepName); existingDep != nil {
						foundStr = fmt.Sprintf(" (Found %s)", existingDep)
					}
					fmt.Printf("- %s%s\n", elementDependency, foundStr)
				}
			}
		}
	}

	if err != nil {
		err = fmt.Errorf("Unable to attach elements related to %s-%s. %s", element.GetType(), element.Name(), err)
		return
	}
	if elementTypeDeps == nil {
		return
	}
	for _, elementDeps := range elementTypeDeps.list {
		for _, elementDependency := range elementDeps {
			elementDepType := elementDependency.GetType()
			elementDepName := elementDependency.Name()

			if existingDep := e.GetElement(elementDepType, elementDepName); existingDep == nil {
				if _, err = e.AddElement(elementDependency); err != nil {
					return
				}

				// Set this elementDependency added as required by element
				elementDependency = e.GetElement(elementDepType, elementDepName)
				element.AddDependencyTo(elementDependency)
			} else {
				var v1, v2 *goversion.Version
				v, _ := existingDep.GetVersion()
				v1 = v.Get()
				v, _ = elementDependency.GetVersion()
				v2 = v.Get()

				// Ensure this found dependency is attached to the current element
				element.AddDependencyTo(existingDep)

				if v2 == nil || (v1 != nil && v1.Compare(v2) >= 0) {
					gotrace.Trace("No version change for %s. Moving to next dependency.", elementDependency.Name())
					continue
				}
				if _, err := e.UpdateElement(existingDep, elementDependency); err != nil {
					return
				}

				// Refreshing dependencies of each element dependencies
				if gotrace.IsDebugMode() {
					fmt.Printf("Checking %s dependencies between %s and %s\n", existingDep.Name(), v1, v2)
				}
				// Update dependencies tree if needed.
				e.RefreshDependencies(existingDep, elementDependency)
			}
		}
	}

	// Cleanup old dependencies
	for depName, depElement := range element.GetDependencies() {
		if de := elementTypeDeps.GetElement(depElement.GetType(), depName); de == nil {
			element.RemoveDependencyTo(depElement)
		}
	}

	if gotrace.IsDebugMode() {
		fmt.Printf("%s has NOW following deps:\n", element)
		element.GetDependencies().DisplayDependencies()
	}
	return
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
		element.Merge(e, existing, newestPolicy)
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
