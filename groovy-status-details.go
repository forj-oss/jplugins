package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"
	"path"

	"github.com/forj-oss/forjj-modules/trace"
)

type groovyStatusDetails struct {
	name       string
	newMd5     string
	oldMd5     string
	sourcePath string
}

func newGroovyStatusDetails(name, sourcePath string) (ret *groovyStatusDetails) {
	ret = new(groovyStatusDetails)
	ret.name = name
	ret.sourcePath = sourcePath
	return
}

func (gsd *groovyStatusDetails) computeM5Sum(bNew bool) (_ bool) {
	groovyFile := path.Join(gsd.sourcePath, gsd.name+".groovy")
	fd, err := os.Open(groovyFile)
	if err != nil {
		gotrace.Error("Unable to read '%s'. %s", groovyFile, err)
		return
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	hash := md5.New()

	if _, err := io.Copy(hash, reader); err != nil {
		gotrace.Error("Unable to generate md5sum data. %s", err)
		return
	}
	md5Data := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	if bNew {
		gsd.newMd5 = md5Data
	} else {
		gsd.oldMd5 = md5Data
	}

	return true
}
