package coremgt

import (
	goversion "github.com/hashicorp/go-version"
)

type pluginJson struct {
	Name        string
	ShortName   string
	Description string
	OldVersion  string
	NewVersion  string
}

// IsNewer return true if the element identify a newer version
func IsUpdated(p pluginJson) (_ bool) {
	if p.OldVersion == "" {
		if p.NewVersion != "" {
			return true
		}
		return
	} else if p.NewVersion == "" {
		return true
	} else {
		oldVersion, _ := goversion.NewVersion(p.OldVersion)
		newVersion, _ := goversion.NewVersion(p.NewVersion)
		return newVersion.GreaterThan(oldVersion)
	}
}