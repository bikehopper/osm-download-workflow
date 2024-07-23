package main

import (
	"errors"
	"fmt"
	"os"

	osm_download_workflow "github.com/bikehopper/osm-download-workflow/src"
)

func main() {
	var argsWithoutProg []string
	if len(os.Args) != 2 {
		panic(errors.New("only accepts one arguemnt"))
	} else {
		argsWithoutProg = os.Args[1:]
	}

	switch argsWithoutProg[0] {
	case "create":
		osm_download_workflow.Create()
	case "worker":
		osm_download_workflow.Worker()
	default:
		fmt.Printf("Must pass 'create', 'worker'\n")
	}
}
