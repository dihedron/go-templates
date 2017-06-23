package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/camelcase"
)

// Node describes a node in the XML tree.
type Node struct {
	xml       interface{}
	container bool
}

var pattern *regexp.Regexp

func init() {
	pattern, _ = regexp.Compile(`^{{[^}}]*}}$`)
}

// tab creates a string with the given number of tabs; each tab has a size of
// two blank spaces in the current implementation.
func tab(count int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", count*2), "")
}

// templatise returns the name of the template parameter for a given tag, e.g.
// <doSomething> becomes DoSomething
func templatise(tag string) string {
	var tokens []string
	if strings.Contains(tag, ".") {
		for _, token := range strings.Split(tag, ".") {
			tokens = append(tokens, strings.Title(token))
		}
	} else {
		tokens = camelcase.Split(tag)
		tokens = append([]string{strings.Title(tokens[0])}, tokens[1:]...)
	}
	return strings.Join(tokens, "")
}
