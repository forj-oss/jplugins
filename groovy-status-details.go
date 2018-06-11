package main

import (
	"bufio"
	"crypto/md5"
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
	descFile := path.Join(gsd.sourcePath, gsd.name, gsd.name+".desc")
	fd, err := os.Open(descFile)
	if err != nil {
		gotrace.Error("Unable to read ")
		return
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	hash := md5.New()

	if _, err := io.Copy(hash, reader); err != nil {
		gotrace.Error("Unable to generate md5sum data. %s", err)
		return
	}

	if bNew {
		gsd.newMd5 = string(hash.Sum(nil))
	} else {
		gsd.oldMd5 = string(hash.Sum(nil))
	}

	return true
}
