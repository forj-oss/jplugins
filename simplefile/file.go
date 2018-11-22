package simplefile

import (
	"fmt"
	"os"
	"sort"
	"strings"
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

// Add fields to a simple data line
func (s *SimpleFile) Add(index int, data ...string) {
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

// WriteSimpleSortedFile save the Simple file from data loaded
func (s *SimpleFile) WriteSimpleSortedFile(sep string) (_ error) {

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
