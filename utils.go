package main

import (
	"fmt"
	"regexp"

	"github.com/fatih/color"
)

// Node describes a node in the XML tree.
type Node struct {
	xml       interface{}
	container bool
}

// Colorf is the type of string formatting functions.
type Colorf func(a ...interface{}) string

var (
	pattern *regexp.Regexp
	bold    Colorf
	green   Colorf
	red     Colorf
)

func init() {
	pattern, _ = regexp.Compile(`^{{[^}}]*}}$`)
	bold = color.New(color.FgWhite, color.Bold).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
	red = color.New(color.FgRed).SprintFunc()
}

// tab creates a string with the given number of tabs; each tab has a size of
// two blank spaces in the current implementation.
func tab(count int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", count*2), "")
}
