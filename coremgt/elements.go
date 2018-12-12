package coremgt

import (
	"fmt"
	"sort"
	"strings"
)

// Elements represents the list of Elements (plugins, features, etc...)
type Elements map[string]Element

// DisplayDependencies display the list of dependencies of that element.
func (l Elements) DisplayDependencies() {
	list := make([]string, 0, len(l))

	for _, element := range l {
		version, _ := element.GetVersion()
		list = append(list, element.GetType()+":"+element.Name()+":"+version.Get().Original())
	}
	fmt.Printf("%d Deps:\n", len(list))
	sort.Strings(list)

	fmt.Printf(" * %s\n", strings.Join(list, "\n * "))
}
