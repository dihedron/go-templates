package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
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
       |  templatise ver. 1.0.0 - Copyright © Andrea Funtò 2017  |
       +---------------------------------------------------------+
	   
A utility to generate the Golang template of a Jenkins job configuration file 
interactively for use as input to the Terraform Jenkins provider.

usage:
  $> templatise <config.xml> <config.tpl> <params.tf>
where:
  config.xml [in]  is the original, non-generic Jenkins job configuration file
  config.tpl [out] is the templatised version of the same file 
  params.tf  [out] is an example of Terraform file defining the job's traits
`
)

// templatise <config.xml> <config.tpl> <params.tf>
func main() {
	if len(os.Args) != 4 {
		fmt.Println(usage)
		os.Exit(1)
	}

	file, _ := os.Open(os.Args[1])
	d := xml.NewDecoder(file)
	stack := NewStack()
	var data string
	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF && token == nil {
				fmt.Printf("done reading the input XML")
				break
			}
			fmt.Printf("error reading token %v: %v\n", token, err)
			break
		}
		switch token := token.(type) {
		case xml.StartElement:
			stack.Push(token.Name.Local)
			data = ""
			//fmt.Printf("START: <%s>\n", token.Name.Local)
		case xml.CharData:
			chars := strings.TrimSpace(string(token.Copy()))
			if len(chars) > 0 {
				//fmt.Printf("CHAR : %q\n", data)
				data = chars
			}
		case xml.EndElement:
			top := stack.Pop()
			if len(data) > 0 {

			}
			fmt.Printf("END  : </%s> (top: %s)\n", token.Name.Local, top.(string))
		case xml.Comment:
			// ignore
		default:
			fmt.Printf("token: %v\n", token)
		}
	}
}

/*
func main() {
	// create a new template and parse the pipeline into it
	t, err := template.New("template").Parse(configuration)
	if err != nil {
		log.Printf("error parsing template: %v\n", err)
		os.Exit(1)
	}

	job := Job{
		Name:        "JobName",
		Description: "This is the description",
		Parameters: map[string]interface{}{
			"KeepDependencies":              true,
			"GitLabConnection":              "http://gitlab.bancaditalia.it/my-project/myproj.git",
			"TriggerOnPush":                 true,
			"TriggerOnMergeRequest":         true,
			"TriggerOpenMergeRequestOnPush": "never",
			"TriggerOnNoteRequest":          true,
			"NoteRegex":                     "Jenkins please retry a build",
			"CISkip":                        true,
			"SkipWorkInProgressMergeRequest": true,
			"SetBuildDescription":            true,
			"BranchFilterType":               "All",
			"SecretToken":                    "{AQAAABAAAAAQwt1GRY9q3ZVQO3gt3epgTsk5dMX+jSacfO7NOzm5Eyk=}",
		},
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, job)
	if err != nil {
		log.Printf("error executing template: %v\n", err)
	}

	fmt.Println("-------------------------------------------------------------------------------------")
	fmt.Printf("%s\n", buffer.String())
	fmt.Println("-------------------------------------------------------------------------------------")
}
*/
