package coremgt

import (
	"fmt"

	"github.com/foize/go.fifo"
)

type elementsCollection struct {
	context             *ElementsType
	orderedElementsType map[string][]Element
	treatedElementsType map[string]Elements
}

func newElementsCollection(context *ElementsType) (ret *elementsCollection) {
	if context == nil {
		return
	}
	ret = new(elementsCollection)
	ret.context = context
	ret.orderedElementsType = make(map[string][]Element)
	ret.treatedElementsType = make(map[string]Elements)
	for elementType, elements := range context.list {
		ret.orderedElementsType[elementType] = make([]Element, 0, len(elements))
		ret.treatedElementsType[elementType] = make(Elements)
	}
	return
}

// BuildOrder create the list in appropriate deps tree order to set versions of elements
func (c *elementsCollection) BuildOrder() (_ map[string][]Element, _ error) {
	if c == nil {
		return nil, nil
	}

	for elementType, elements := range c.context.list {
		queue := fifo.NewQueue()
		orderedElements := c.orderedElementsType[elementType]
		treatedElements := c.treatedElementsType[elementType]
		for name, element := range elements {
			if element.IsFixed() {
				treatedElements[name] = element
				continue
			}
			queue.Add(element)
		}

		loopDetect := 0
		loopQueue := queue.Len()
		for queue.Len() > 0 {
			element := queue.Next().(Element)

			c.orderedElementsType[elementType] = orderedElements
			c.treatedElementsType[elementType] = treatedElements
			if c.treatElementOrder(queue, element) {
				orderedElements = append(orderedElements, element)
				treatedElements[element.Name()] = element
			} else {
				queue.Add(element)
			}
			if loopQueue == queue.Len() {
				loopDetect++
			} else {
				loopDetect = 0
			}
			if loopDetect > queue.Len() {
				return nil, fmt.Errorf("Unable to organize list of elements. Looping in dependencies")
			}
		}
	}
	return c.orderedElementsType, nil
}

// treatElementOrder determine if all dependencies are treated or not
// if all treated, return true
// else return false
func (c *elementsCollection) treatElementOrder(queue *fifo.Queue, element Element) (treated bool) {
	if c == nil {
		return
	}
	treated = true
	switch element.GetType() {
	case pluginType:
		var refPlugin *RepositoryPlugin
		if version, _ := element.GetVersion() ; version.Get() == nil {
			refPlugin, _ = c.context.ref.Get(element.Name()) // Get latest by default
		} else {
			refPlugin, _ = c.context.ref.Get(element.Name(), version.Get().Original())
		}

		if refPlugin == nil {
			// No information about dependencies from updates. Ignoring it and
			// consider it with no dependencies. So it return true.
			return true
		}

		treatedElements, found := c.treatedElementsType[element.GetType()]
		if !found {
			return
		}

		for _, dep := range refPlugin.Dependencies {
			if _, found = treatedElements[dep.Name]; !found {
				return false
			}
		}

	}
	return
}
