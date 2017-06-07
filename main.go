package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
)

type Job struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

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
