package utils

import (
	"os"
	"path"
)

func CheckPath(aPath string) (_ bool) {
	if info, err := os.Stat(aPath) ; err != nil {
		return
	} else if !info.IsDir() {
		return
	}
	return true
}

// CheckFile simply verify if the file exist.
func CheckFile(filepath, file string) (_ bool) {
	simpleFile := path.Join(filepath, file)
	if info, err := os.Stat(simpleFile); err != nil {
		return
	} else if info.IsDir() {
		return
	}
	return true
}