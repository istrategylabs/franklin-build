package main

import (
	"bytes"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/istrategylabs/franklin-build/logging"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"math/rand"
	"net/http"
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
	ENV_ID     int    `json:"environment_id" binding:"required"`
}

type config struct {
	BUILDLOCATION  string
	FRANKLINAPIURL string
	DEPLOYROOTPATH string
	ENV            string
}

var Config config

// Investigate a better way to do this
func init() {
	Config.BUILDLOCATION = os.Getenv("BUILD_LOCATION")
	if Config.BUILDLOCATION == "" {
		logging.LogToFile("Missing environment variable BUILD_LOCATION")
		panic("Missing environment variable BUILD_LOCATION")
	}
	Config.FRANKLINAPIURL = os.Getenv("API_URL")
	if Config.FRANKLINAPIURL == "" {
		logging.LogToFile("Missing environment variable API_URL")
		panic("Missing environment variable API_URL")
	}
	Config.DEPLOYROOTPATH = os.Getenv("DEPLOY_ROOT_FOLDER")
	if Config.DEPLOYROOTPATH == "" {
		logging.LogToFile("Missing environment variable DEPLOY_ROOT_FOLDER")
		panic("Missing environment variable DEPLOY_ROOT_FOLDER")
	}
	Config.ENV = os.Getenv("ENV")

}

func makePutRequest(dockerInfo DockerInfo, data string) {

	if Config.ENV == "test" {
		return
	}

	url := apiUrl(dockerInfo)

	var jsonStr = []byte(fmt.Sprintf(`{"status":"%s"}`, data))
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	logging.LogToFile(err)

	defer resp.Body.Close()
	logging.LogToFile(resp.Status)

}

func apiUrl(dockerInfo DockerInfo) string {
	url := Config.FRANKLINAPIURL + "/build/" + strconv.Itoa(dockerInfo.ENV_ID) + "/update/"
	return url

}

// BuildDockerContainer executes a docker build command and assigns it a random tag
func BuildDockerContainer(com chan string) {
	// First we will seed the random number generator
	rand.Seed(time.Now().UnixNano())
	randomTag := strconv.Itoa(rand.Intn(1000))
	// TODO: REMOVE TEMP LAYERS
	out, err := exec.Command("docker", "build", "--no-cache=True", "-t", randomTag, ".").Output()
	if err != nil {
		logging.LogToFile("There was an error building the docker container")
	}
	logging.LogToFile(string(out))
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

func GenerateDockerFile(dockerInfo DockerInfo, buildDir string) error {
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
func GrabBuiltStaticFiles(dockerImageID, projectName, transferLocation string) {
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

func Build(buildDir, projectName, projectPath string) string {
	c1 := make(chan string)
	go BuildDockerContainer(c1)

	// Looping until we get notification on the channel c1 that the build has finished
	for {
		select {
		case buildTag := <-c1:
			logging.LogToFile("Container built...transfering built files...")
			GrabBuiltStaticFiles(buildTag, projectName, buildDir)
			if Config.ENV != "test" {
				rsyncProject(buildDir+"/public/*", Config.DEPLOYROOTPATH+projectPath)
			}
			return "success"
		}
	}
}

func rsyncProject(buildDir, remoteLoc string) {
	rsyncCommand := exec.Command("scripts/rsync_project.sh", buildDir, remoteLoc)
	res, err := rsyncCommand.CombinedOutput()
	logging.LogToFile(err)
	logging.LogToFile(string(res))
}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	logging.LogToFile(fmt.Sprintf("Started building %s", dockerInfo.REPO_NAME))
	err := GenerateDockerFile(dockerInfo, ".")

	if err != nil {
		// this line can probably be removed but we need to figure what if anything we should return in responses
		r.JSON(500, map[string]interface{}{"success": "false"})
		makePutRequest(dockerInfo, "FAL")

	}

	logging.LogToFile("Dockerfile generated successfully...building container...")

	go Build(Config.BUILDLOCATION, dockerInfo.REPO_NAME, dockerInfo.PATH)
	// this line can probably be removed but we need to figure what if anything we should return in responses
	r.JSON(200, map[string]interface{}{"success": "true"})
	// We are assuming the build is successful for now.
	makePutRequest(dockerInfo, "SUC")

}
