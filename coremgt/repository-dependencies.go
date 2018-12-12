package coremgt

import (
	goversion "github.com/hashicorp/go-version"
)

// RepositoryDependencies is a array of dependencies
type RepositoryDependencies []RepositoryDependency

// GetVersion return the dependency version of a given plugin name.
func (r RepositoryDependencies) GetVersion(name string) *goversion.Version {

	for _, plugin := range r {
		if plugin.Name == name {
			v, _ := goversion.NewVersion(plugin.Version)
			return v
		}
	}
	return nil
}
