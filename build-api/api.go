package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"os"
	"os/exec"
	"text/template"
)

type DockerInfo struct {
	DEPLOY_KEY string `json:"deploy_key" binding:"required"`
	BRANCH     string `json:"branch"`
	TAG        string `json:"tag"`
	HASH       string `json:"git_hash" binding:"required"`
	REPO_OWNER string `json:"repo_owner" binding:"required"`
	PATH       string `json:"path" binding:"required"`
	REPO_NAME  string `json:"repo_name" binding:"required"`
}

func buildDockerContainer() {
	// We can pass in a callback here, or just handle the status update
	// request from this function
	buildCommand := exec.Command("docker", "build", "--no-cache=True", "--tags='franklin_builder_tmp:tmp'", ".")
	if err := buildCommand.Run(); err != nil {
		fmt.Println(os.Stderr, err)
	}

	tearDown := exec.Command("scripts/tear_down_project.sh")
	if err := tearDown.Run(); err != nil {
		fmt.Println(os.Stderr, err)
	}

	os.Remove("tmp/")
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Post("/build", binding.Bind(DockerInfo{}), BuildDockerFile)
	m.Run()

}

func GenerateDockerFile(dockerInfo DockerInfo, buildDir string) {
	tmp_dir := buildDir

	// Create a new Dockerfile template parses template definition
	docker_tmpl, err := template.ParseFiles("templates/dockerfile.tmplt")
	HandleErr(err)

	// Create tmp directory
	err = os.Mkdir(tmp_dir, 0770)
	HandleErr(err)

	// Create file
	f, err := os.Create(tmp_dir + "/Dockerfile")
	HandleErr(err)
	defer f.Close()

	//Apply the Dockerfile template to the docker info from the request
	err = docker_tmpl.Execute(f, dockerInfo)
	HandleErr(err)
}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	GenerateDockerFile(dockerInfo, "tmp")
	go buildDockerContainer()
	r.JSON(200, map[string]interface{}{"success": true})
}
