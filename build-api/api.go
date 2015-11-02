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
	"strings"
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
func buildDockerContainer(com chan string) {
	randomTag := strconv.Itoa(rand.Intn(1000))
	_, err := exec.Command("docker", "build", "--no-cache=True", "-t", randomTag, ".").Output()
	logging.LogToFile(err)

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

// grabBuiltStaticFiles issues a `docker run` command to the container image
// we created that will transfer built files to specified location
func grabBuiltStaticFiles(dockerImageID string, transferLocation string) {
	// This will need to change eventually...hardcoding for testing purposes
	// projectName := "franklin-test"
	mountStringSlice := []string{transferLocation, ":", "tmp_mount"}
	mountString := strings.Join(mountStringSlice, "")
	fmt.Println(mountString)

	// err = os.Mkdir(transferLocation, 0770)
	// logging.LogToFile(err)

	// _, err := exec.Command("docker", "run", "-v", "-t", randomTag, ".").Output()

	// docker run -v `pwd`/build_directory:/tmp_mount tmp_id cp -r /$project_name/dist /tmp_mount
	// logging.LogToFile(err)

}

func build() {
	c1 := make(chan string)
	go buildDockerContainer(c1)

	// Looping until we get notification on the channel c1 that the build has finished
	for {
		select {
		case buildTag := <-c1:
			grabBuiltStaticFiles(buildTag, "tmp_build_dir")
		}
	}
}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	err := generateDockerFile(dockerInfo, ".")

	if err != nil {
		r.JSON(500, map[string]interface{}{"success": false})
	}

	go build()
	r.JSON(200, map[string]interface{}{"success": true})
}
