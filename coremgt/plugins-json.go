package coremgt

import (
	"github.com/robpike/filter"
)

type pluginsJson []pluginJson

func (p pluginsJson) IsLast(index int) bool {
	return index+1 >= len(p)
}

// UpdatedList return a plugins list identifying only new versions.
func (p pluginsJson) Newest() (ret pluginsJson) {
	list := filter.Choose(p, IsUpdated)
	ret = list.(pluginsJson)
	return
}
