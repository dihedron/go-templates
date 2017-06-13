package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"flag"

	"log"

	"path/filepath"

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
	all := flag.Bool("all", false, "write all potential values, even empty ones [default: false]")
	inline := flag.Bool("inline", false, "produce an HCL file with inlined template [default: false]")
	flag.Parse()

	if !*debug {
		log.SetOutput(ioutil.Discard)
	}

	if len(flag.Args()) != 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	var err error
	var hcl, tpl *os.File
	if hcl, tpl, err = openFiles(flag.Args()[0], *inline); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening output files: %v\n", err)
		os.Exit(1)
	}
	defer hcl.Close()

	handler := &Handler{
		stack: stack.New(),
		all:   *all,
		hcl:   bufio.NewWriter(hcl),
	}
	if tpl != nil {
		handler.tpl = bufio.NewWriter(tpl)
		defer tpl.Close()
	}
	defer handler.Close()

	parser := &sax.Parser{
		EventHandler: handler,
		ErrorHandler: handler,
	}

	file, _ := os.Open(flag.Args()[0])
	defer file.Close()
	parser.Parse(file)
}

func openFiles(path string, inline bool) (hcl *os.File, tpl *os.File, err error) {

	dir := filepath.Dir(path)
	hclPath := dir + "/" + filepath.Base(path) + ".hcl"

	if _, err = os.Stat(hclPath); err == nil {
		fmt.Fprintf(os.Stderr, "File %s exists already\n", hclPath)
		err = fmt.Errorf("File %s exists already", hclPath)
		return
	}

	if hcl, err = os.Create(hclPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", hclPath, err)
		return
	}

	if !inline {
		tplPath := dir + "/" + filepath.Base(path) + ".tpl"
		if _, err = os.Stat(tplPath); err == nil {
			fmt.Fprintf(os.Stderr, "File %s exists already\n", tplPath)
			err = fmt.Errorf("File %s exists already", hclPath)
			return
		}
		if tpl, err = os.Create(tplPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", tplPath, err)
			return
		}
	}
	return
}
