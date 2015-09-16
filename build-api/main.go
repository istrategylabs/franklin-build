package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"net/http"
	"os"
	"os/exec"
)

func buildDockerContainer() {
	// We can pass in a callback here, or just handle the status update
	// request from this function
	command := "sudo docker build --no-cache=True --tags='franklin_builder_tmp:tmp' ."
	// 1. Will need to expand this to make sure we are running the command
	// in the style we want
	if err := exec.Command(command).Run(); err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
	// 2. Let's wait for and handle any errors after each shell command
	// 3. Call 'tear_down_project' and remove tmp directory
}

func main() {
	m := martini.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})

	m.Post("/build", func(res http.ResponseWriter, req *http.Request) {
		// 1. Need to parse json from 'req'
		// 2. Need to check that repoName, repoOwner and gitHash exist
		// 3. Compile Dockerfile with template language we will choose
		// 4. This should be all that is needed for concurrent/async builds
		go buildDockerContainer()
		// 5. Return a json response with success or error if present
	})
	m.Run()
}
