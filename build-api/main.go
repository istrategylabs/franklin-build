package main

import (
	"fmt"
	"github.com/codegangsta/martini-contrib/binding"
	"github.com/go-martini/martini"
	"os"
	"os/exec"
	"text/template"
)

type DockerInfo struct {
	BRANCH     string `json:"git_branch" binding:"required"`
	HASH       string `json:"git_hash" binding:"required"`
	REPO_OWNER string `json:"repo_owner" binding:"required"`
	REMOTE_LOC string `json:"path" binding:"required"`
	REPO_NAME  string `json:"repo_name" binding:"required"`
}

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

	m.Post("/build", binding.Bind(DockerInfo{}), func(dockerInfo DockerInfo) string {

		tmp_dir := "tmp"

		docker_tmpl, err := template.ParseFiles("templates/dockerfile.tmplt")
		if err != nil {
			fmt.Println(err)
		}

		err = os.Mkdir(tmp_dir, 0770)
		if err != nil {
			fmt.Println(err)
		}

		f, err := os.Create(tmp_dir + "/Dockerfile")
		if err != nil {
			fmt.Println(err)
		}
		defer f.Close()

		err = docker_tmpl.Execute(f, dockerInfo)
		if err != nil {
			fmt.Println(err)
		}
		go buildDockerContainer()
		return "success"
	})
	m.Run()

}
