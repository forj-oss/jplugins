package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	git "github.com/forj-oss/go-git"

	"github.com/forj-oss/forjj-modules/trace"
)

type groovyStatusDetails struct {
	name          string
	newMd5        string
	oldMd5        string
	newCommit     string
	oldCommit     string
	commitHistory []string
	sourcePath    string
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

func (gsd *groovyStatusDetails) defineVersion(bNew bool) (_ bool) {
	if gsd.commitHistory == nil {
		err := git.RunInPath(gsd.sourcePath, func() (_ error) {
			historyData, err := git.Get("log", "--pretty=%H", gsd.name+".groovy")
			if err != nil {
				return fmt.Errorf("Unable to get file '%s' history from GIT. %s", gsd.name+".groovy", err)
			}
			gsd.commitHistory = strings.Split(strings.Trim(historyData, " \n"), "\n")
			return
		})
		if err != nil {
			gotrace.Error("Unable to define the groovy '%s' version (commit ID)> %s", gsd.name, err)
			return
		}
	}
	if len(gsd.commitHistory) == 0 {
		return true
	}
	latest := gsd.commitHistory[0]
	if bNew {
		gsd.newCommit = latest
	} else {
		gsd.oldCommit = latest
	}
	return true
}

func (gsd *groovyStatusDetails) installIt(destPath string) error {
	git.RunInPath(gsd.sourcePath, func() error {
		if git.Do("checkout", gsd.newCommit) != 0 {
			return fmt.Errorf("Unable to checkout version %s (commit ID) for %s", gsd.newCommit, gsd.name+".groovy")
		}
		return nil
	})
	srcFile := path.Join(gsd.sourcePath, gsd.name+".groovy")
	destFile := path.Join(destPath, path.Base(gsd.name)+".groovy")
	srcfd, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer srcfd.Close()

	var destfd *os.File
	destfd, err = os.Create(destFile)
	if err != nil {
		return err
	}
	defer destfd.Close()

	_, err = io.Copy(destfd, srcfd)
	if err != nil {
		return fmt.Errorf("Unable to copy %s to %s. %s", srcFile, destFile, err)
	}

	gotrace.Trace("Copied: %s => %s", srcFile, destFile)
	return nil
}
