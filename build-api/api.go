package main

import (
	"./logging"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"text/template"
)

// A DockerInfo represents the structure of data coming from Franklin-api
type DockerInfo struct {
	DEPLOY_KEY string `json:"deploy_key" binding:"required"`
	BRANCH     string `json:"branch"`
	TAG        string `json:"tag"`
	HASH       string `json:"git_hash" binding:"required"`
	REPO_OWNER string `json:"repo_owner" binding:"required"`
	PATH       string `json:"path" binding:"required"`
	REPO_NAME  string `json:"repo_name" binding:"required"`
}

// buildDockerContainer executes a docker build command and assigns it a random tag
func buildDockerContainer() string {
	randomTag := strconv.Itoa(rand.Intn(1000))
	_, err := exec.Command("docker", "build", "--no-cache=True", "-t", randomTag, ".").Output()

	logging.LogToFile(err)

	// tearDown := exec.Command("scripts/tear_down_project.sh")
	// if err := tearDown.Run(); err != nil {
	// 	fmt.Println(os.Stderr, err)
	// }

	// os.Remove("tmp/")
	return randomTag
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())

	// Simple 'health' endpoint for AWS load-balancer health checks
	m.Get("/health", func() string {
		return "Hello world!"
	})

	m.Post("/build", binding.Bind(DockerInfo{}), BuildDockerFile)
	m.Run()
}

func GenerateDockerFile(dockerInfo DockerInfo, buildDir string) error {
	var err_return error

	// Create a new Dockerfile template parses template definition
	docker_tmpl, err := template.ParseFiles("templates/dockerfile.tmplt")
	logging.LogToFile(err)
	err_return = err

	// Create tmp directory
	err = os.Mkdir(buildDir, 0770)
	logging.LogToFile(err)
	err_return = err

	// Create file
	f, err := os.Create(buildDir + "/Dockerfile")
	logging.LogToFile(err)
	err_return = err
	defer f.Close()

	//Apply the Dockerfile template to the docker info from the request
	err = docker_tmpl.Execute(f, dockerInfo)
	err_return = err
	logging.LogToFile(err)

	return err_return
}

func GrabBuiltStaticFiles(dockerImageID string, transferLocation string) {

}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	err := GenerateDockerFile(dockerInfo, ".")

	if err != nil {
		r.JSON(500, map[string]interface{}{"success": false})
	}

	go buildDockerContainer()
	r.JSON(200, map[string]interface{}{"success": true})
}
