package main

import (
	goversion "github.com/hashicorp/go-version"
)

type versionStruct struct {
	original string
	version *goversion.Version
}

func (v versionStruct) String() string {
	return v.original
}

func (v *versionStruct) Set(value string) (err error) {
	if v == nil {
		return
	}
	v.original = value
	if value != "new" {
		v.version, err = goversion.NewVersion(value)
	} else {
		v.version = nil
	}
	return
}

func (v versionStruct) Get() (ret *goversion.Version) {
	return v.version
}