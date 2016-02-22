package main

import (
	"bytes"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/istrategylabs/franklin-build/logging"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"io/ioutil"
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
	DEPLOY_KEY string `json:"deploy_key" binding:"required"`     // SSH Key used to clone private repos
	BRANCH     string `json:"branch"`                            // Branch to deploy
	TAG        string `json:"tag"`                               // Tag to deploy
	HASH       string `json:"git_hash" binding:"required"`       // Commit hash
	REPO_OWNER string `json:"repo_owner" binding:"required"`     // ID of Github repo owner
	PATH       string `json:"path" binding:"required"`           // "Qualified" githubOrganizationID/commitHash
	REPO_NAME  string `json:"repo_name" binding:"required"`      // Name of Github repository
	ENV_ID     int    `json:"environment_id" binding:"required"` // Environment ID from API (staging, production etc)
}

type config struct {
	BUILDLOCATION  string // Location of built files on build server
	FRANKLINAPIURL string // URL of Franklin-API
	DEPLOYROOTPATH string // The root of where we want to rsync files to
	ENV            string // Environment of running Franklin-build instance
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
func BuildDockerContainer(com, quit chan string, buildServerPath string) {
	// First we will seed the random number generator
	rand.Seed(time.Now().UnixNano())
	randomTag := strconv.Itoa(rand.Intn(1000))
	// TODO: REMOVE TEMP LAYERS
	out, err := exec.Command("docker", "build", "--no-cache=True", "-t", randomTag, buildServerPath).Output()
	logging.LogToFile(string(out))
	if err != nil {
		quit <- "fail"
	}
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

func GenerateDockerFile(dockerInfo DockerInfo, buildServerPath string) error {
	var err_return error

	// Create a new Dockerfile template parses template definition
	docker_tmpl, err := template.ParseFiles("templates/dockerfile.tmplt")
	logging.LogToFile(err)
	err_return = err

	f, err := os.Create(buildServerPath + "/Dockerfile")
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
func GrabBuiltStaticFiles(dockerImageID, projectName, buildServerPath string) {
	// Not sure if this is the best way to handle "dynamic strings"
	mountStringSlice := []string{buildServerPath, ":", "/tmp_mount"}
	mountString := strings.Join(mountStringSlice, "")

	transfer := exec.Command("scripts/transfer_files.sh", mountString, dockerImageID, projectName)
	res, err := transfer.CombinedOutput()
	logging.LogToFile(err)
	logging.LogToFile(string(res))
}

func Build(buildServerPath string, dockerInfo DockerInfo) string {
	c1 := make(chan string)
	quit := make(chan string)
	go BuildDockerContainer(c1, quit, buildServerPath)
	makePutRequest(dockerInfo, "BLD")

	// Looping until we get notification on the channel c1 that the build has finished
	for {
		select {
		case <-quit:
			logging.LogToFile("There was an error building the docker container")
			makePutRequest(dockerInfo, "FAL")
			return "fail"
		case buildTag := <-c1:
			logging.LogToFile("Container built...transfering built files...")
			GrabBuiltStaticFiles(buildTag, dockerInfo.REPO_NAME, buildServerPath)
			if Config.ENV != "test" {
				rsyncProject(buildServerPath+"/public/*", Config.DEPLOYROOTPATH+dockerInfo.PATH)
			}
			makePutRequest(dockerInfo, "SUC")
			return "success"
		}
	}
}

func rsyncProject(buildServerPath, remoteLoc string) {
	rsyncCommand := exec.Command("scripts/rsync_project.sh", buildServerPath, remoteLoc)
	res, err := rsyncCommand.CombinedOutput()
	logging.LogToFile(err)
	logging.LogToFile(string(res))
}

func createTempSSHKey(dockerInfo DockerInfo, buildServerPath string) error {
	var err_return error

	d1 := []byte(dockerInfo.DEPLOY_KEY)
	err := ioutil.WriteFile(buildServerPath+"/id_rsa", d1, 0644)
	logging.LogToFile(err)
	err_return = err

	return err_return
}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	logging.LogToFile(fmt.Sprintf("Started building %s", dockerInfo.REPO_NAME))

	// Let's hold a reference to the project's path to build on
	buildServerPath := Config.BUILDLOCATION + "/" + dockerInfo.PATH

	err := os.MkdirAll(buildServerPath, 0770)
	logging.LogToFile(err)

	err = createTempSSHKey(dockerInfo, buildServerPath)
	logging.LogToFile(err)
	err = GenerateDockerFile(dockerInfo, buildServerPath)
	logging.LogToFile(err)

	if err != nil {
		// this line can probably be removed but we need to figure what if anything we should return in responses
		r.JSON(500, map[string]interface{}{"success": "false"})
		makePutRequest(dockerInfo, "FAL")

	}

	logging.LogToFile("Dockerfile generated successfully...building container...")

	// Need to pass in more informabout about the location of
	go Build(buildServerPath, dockerInfo)

	// this line can probably be removed but we need to figure what if anything we should return in responses
	r.JSON(200, map[string]interface{}{"success": "true"})

}
