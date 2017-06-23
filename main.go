package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
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
  $> jted [-include-empty-values] [-embed-template] <config.xml>
where:
  -include-empty-values
    specifies whether empty tags in the original config.xml should be used
	to generate values, that is <tag></tag> or <tag/> will become
	<tag>{{- Tag -}}</tag> [default: false, that is omit empty tags]
  -embed-template
    specifies whether the generated config.xml template should be embedded 
	in the generated HCL (.tf) file as a template field [default: false]
  config.xml [in]  is the original, non-generic Jenkins job configuration file
`
)

// jted <config.xml> <config.tpl> <params.tf>
func main() {

	includeEmptyValues := flag.Bool("include-empty-values", false, "write all potential values, even empty ones [default: false]")
	embedTemplate := flag.Bool("embed-template", false, "produce an HCL file with inlined template [default: false]")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	handler := &Handler{
		IncludeEmptyValues: *includeEmptyValues,
		EmbedConfigXML:     *embedTemplate,
		stack:              stack.New(),
		parameters:         map[string]string{},
	}

	parser := &sax.Parser{
		EventHandler: handler,
		ErrorHandler: handler,
	}

	file, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatalf("Error opening input file: %v", err)
	}
	defer file.Close()
	err = parser.Parse(file)
	if err != nil {
		log.Fatalf("Error parsing input file: %v", err)
	}

	hcl, err := openFile(getHCLFileName(flag.Args()[0]))
	if err != nil {
		log.Fatalf("Error opening HCL for writing: %v", err)
	}
	defer hcl.Close()

	hclWriter := bufio.NewWriter(hcl)
	if handler.EmbedConfigXML {
		// if the template is embedded, then only the HCL should be written out as is,
		// it contains everything already
		hclWriter.Write(handler.HCL.Bytes())
		hclWriter.Flush()
	} else {
		// if the tempate is not embedded, the HCL must be written out after replacing
		// the name of the tpl file (a "%s" was left by the handler in the buffer for
		// this purpose)...
		hclWriter.WriteString(fmt.Sprintf(handler.HCL.String(), getConfigXMLTemplateFileName(flag.Args()[0])))
		hclWriter.Flush()

		// ... and then the template must be written out too to its own writer
		tpl, err := openFile(getConfigXMLTemplateFileName(flag.Args()[0]))
		if err != nil {
			log.Fatalf("Error opening config.xml template file for writing: %v", err)
		}
		defer tpl.Close()
		tplWriter := bufio.NewWriter(tpl)

		tplWriter.Write(handler.ConfigXML.Bytes())
		tplWriter.Flush()
	}
}

func getHCLFileName(configXML string) string {
	return filepath.Dir(configXML) + "/" + filepath.Base(configXML) + ".hcl"
}

func getConfigXMLTemplateFileName(configXML string) string {
	return filepath.Dir(configXML) + "/" + filepath.Base(configXML) + ".tpl"
}

func openFile(path string) (file *os.File, err error) {
	if _, err = os.Stat(path); err == nil {
		fmt.Fprintf(os.Stderr, "File %s exists already\n", path)
		err = fmt.Errorf("File %s exists already", path)
		return
	}
	if file, err = os.Create(path); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", path, err)
		return
	}
	return
}
