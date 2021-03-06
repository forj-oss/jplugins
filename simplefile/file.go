package simplefile

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
)

// SimpleFile root struct to manage a Simple file format separated by colon.
type SimpleFile struct {
	file string
	cols int
	data map[string]Line
}

// Line describe the data split from a line.
type Line struct {
	data []string
}

// NewSimpleFile return a struct of SimpleFile as defined here
func NewSimpleFile(filename string, cols int) (ret *SimpleFile) {
	ret = new(SimpleFile)
	ret.file = filename
	ret.cols = cols
	ret.data = make(map[string]Line)
	return
}

// AddWithKeyIndex add fields to a simple data line and set key as field value (index)
func (s *SimpleFile) AddWithKeyIndex(index int, data ...string) {
	if len(data) == 0 {
		return
	}
	if index >= len(data) {
		index = 0
	}
	cols := make([]string, 0, s.cols)
	for curIndex, field := range data {
		if curIndex >= s.cols {
			break
		}
		cols = append(cols, field)
	}

	s.data[data[index]] = Line{data: cols}
}

// AddWithKeyString add fields to a simple data line and index them thanks to the key string.
func (s *SimpleFile) AddWithKeyString(key string, data ...string) {
	if len(data) == 0 {
		return
	}
	if key == "" {
		key = data[0]
	}
	cols := make([]string, 0, s.cols)
	for curIndex, field := range data {
		if curIndex >= s.cols {
			break
		}
		cols = append(cols, field)
	}

	s.data[key] = Line{data: cols}
}

// WriteSorted save the Simple file from data loaded
func (s *SimpleFile) WriteSorted(sep string) (_ error) {

	if sep == "" {
		sep = ":"
	}

	list := make([]string, len(s.data))

	iCount := 0
	for name := range s.data {
		list[iCount] = name
		iCount++
	}

	sort.Strings(list)

	fd, err := os.OpenFile(s.file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("Unable to write '%s'. %s", s.file, err)
	}
	defer fd.Close()

	for _, name := range list {
		fmt.Fprintln(fd, strings.Join(s.data[name].data, sep))
	}

	return
}

// readFromSimpleFormat read a simple description file for plugins or groovies.
func (s *SimpleFile) Read(sep string, treatData func([]string) (error)) (_ error) {
	fd, err := os.Open(s.file)
	if err != nil {
		return fmt.Errorf("Unable to open file '%s'. %s", s.file, err)
	}

	if gotrace.IsDebugMode() {
		fmt.Printf("Reading %s\n--------\n", s.file)
	}
	defer func() {
		fd.Close()
		if gotrace.IsDebugMode() {
			fmt.Printf("-------- %s closed.\n", s.file)
		}
	}()

	scanFile := bufio.NewScanner(fd)

	for scanFile.Scan() {
		line := scanFile.Text()
		line = strings.Trim(line, " ")

		if line == "" || line[0] == '#' {
			continue
		}

		if gotrace.IsDebugMode() {
			fmt.Printf("%s: == %s ==\n", path.Base(s.file), line)
		}
		pluginRecord := strings.Split(line, sep)

		treatData(pluginRecord)
	}
	return
}
