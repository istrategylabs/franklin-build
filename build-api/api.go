package main

import (
	"./logging"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"
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
func buildDockerContainer(com chan string) {
	// First we will seed the random number generator
	rand.Seed(time.Now().UnixNano())
	randomTag := strconv.Itoa(rand.Intn(1000))
	exec.Command("docker", "build", "--no-cache=True", "-t", randomTag, ".").Run()

	// Passing along the randomTag associated with the built docker container to the channel 'com'
	com <- randomTag
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

func generateDockerFile(dockerInfo DockerInfo, buildDir string) error {
	var err_return error

	// Create a new Dockerfile template parses template definition
	docker_tmpl, err := template.ParseFiles("templates/dockerfile.tmplt")
	logging.LogToFile(err)
	err_return = err

	err = os.Mkdir(buildDir, 0770)
	logging.LogToFile(err)
	err_return = err

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

// grabBuiltStaticFiles issues a `docker run` command to the container image
// we created that will transfer built files to specified location
func grabBuiltStaticFiles(dockerImageID string, transferLocation string) {
	// This will need to change eventually...hardcoding for testing purposes
	projectName := "franklin-test"

	// Not sure if this is the best way to handle "dynamic strings"
	mountStringSlice := []string{transferLocation, ":", "/tmp_mount"}
	mountString := strings.Join(mountStringSlice, "")

	err := os.Mkdir(transferLocation, 0770)
	logging.LogToFile(err)

	transfer := exec.Command("scripts/transfer_files.sh", mountString, dockerImageID, projectName)
	res, err := transfer.CombinedOutput()
	logging.LogToFile(err)
	logging.LogToFile(string(res))
}

func build(buildDir string) {
	c1 := make(chan string)
	go buildDockerContainer(c1)

	// Looping until we get notification on the channel c1 that the build has finished
	for {
		select {
		case buildTag := <-c1:
			logging.LogToFile("Container built...transfering built files")
			grabBuiltStaticFiles(buildTag, buildDir)
		}
	}
}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	err := generateDockerFile(dockerInfo, ".")

	if err != nil {
		r.JSON(500, map[string]interface{}{"success": false})
	}

	logging.LogToFile("Dockerfile generated successfully...building container...")
	// TODO: Obviously change this
	go build("/Users/gindi/Desktop/tmp_build_dir")
	r.JSON(200, map[string]interface{}{"success": true})
}
