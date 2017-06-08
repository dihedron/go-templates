package main

import (
	"fmt"
	"os"

	"github.com/dihedron/jted/sax"
	"github.com/dihedron/jted/stack"
)

/*
type Job struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}
*/

const (
	usage = `
       +---------------------------------------------------------+
       |     JTed ver. 1.0.0 - Copyright © Andrea Funtò 2017     |
       +---------------------------------------------------------+
	   
A utility to generate interactively the Golang template of a Jenkins job 
configuration file, for use as input to the Terraform Jenkins provider.

usage:
  $> jted <config.xml> <config.tpl> <params.tf>
where:
  config.xml [in]  is the original, non-generic Jenkins job configuration file
  config.tpl [out] is the templatised version of the same job 
  params.tf  [out] is an example of Terraform file defining the job's traits
`
)

// jted <config.xml> <config.tpl> <params.tf>
func main() {
	if len(os.Args) != 4 {
		fmt.Println(usage)
		os.Exit(1)
	}

	handler := &Handler{
		stack: stack.New(),
	}

	file, _ := os.Open(os.Args[1])
	parser := &sax.Parser{
		EventHandler: handler,
		ErrorHandler: handler,
	}

	parser.Parse(file)
}
