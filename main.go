package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"flag"

	"log"

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
  $> jted [-debug] <config.xml>
where:
  -debug specifies whether the log messages should be written [default: false]
  config.xml [in]  is the original, non-generic Jenkins job configuration file
`
)

// jted <config.xml> <config.tpl> <params.tf>
func main() {

	debug := flag.Bool("debug", false, "whether debug messages should be written out to STDERR")
	flag.Parse()

	if !*debug {
		log.SetOutput(ioutil.Discard)
	}

	if len(flag.Args()) != 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	handler := &PrintHandler{
		stack: stack.New(),
	}

	file, _ := os.Open(flag.Args()[0])
	defer file.Close()
	parser := &sax.Parser{
		EventHandler: handler,
		ErrorHandler: handler,
	}

	parser.Parse(file)
}
