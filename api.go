package main

import (
	"bytes"
	"fmt"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/go-martini/martini"
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
	DEPLOY_KEY  string `json:"deploy_key" binding:"required"`  // SSH Key used to clone private repos
	BRANCH      string `json:"branch" binding:"required"`      // Branch to deploy
	HASH        string `json:"git_hash" binding:"required"`    // Commit hash
	REPO_OWNER  string `json:"repo_owner" binding:"required"`  // ID of Github repo owner
	PATH        string `json:"path" binding:"required"`        // "Qualified" githubOrganizationID/commitHash
	REPO_NAME   string `json:"repo_name" binding:"required"`   // Name of Github repository
	ENVIRONMENT string `json:"environment" binding:"required"` // Environment name (staging, production etc)
	CALLBACK    string `json:"callback" binding:"required"`    // The webhook callback for api with results
}

type config struct {
	BUILDLOCATION  string // Location of built files on build server
	DEPLOYROOTPATH string // The root of where we want to rsync files to
	ENV            string // Environment of running Franklin-build instance
}

var Config config

// Investigate a better way to do this
func init() {
	log.SetHandler(cli.New(os.Stdout))

	Config.BUILDLOCATION = os.Getenv("BUILD_LOCATION")
	if Config.BUILDLOCATION == "" {
		log.Warn("Missing environment variable BUILD_LOCATION")
		panic("Missing environment variable BUILD_LOCATION")
	}
	Config.DEPLOYROOTPATH = os.Getenv("DEPLOY_ROOT_FOLDER")
	if Config.DEPLOYROOTPATH == "" {
		log.Warn("Missing environment variable DEPLOY_ROOT_FOLDER")
		panic("Missing environment variable DEPLOY_ROOT_FOLDER")
	}
	Config.ENV = os.Getenv("ENV")

}

func updateApiStatus(dockerInfo DockerInfo, data string) {

	ctx := log.WithFields(log.Fields{
		"repo": dockerInfo.REPO_NAME,
		"env":  dockerInfo.ENVIRONMENT,
	})

	if Config.ENV == "test" {
		return
	}

	var jsonStr = []byte(fmt.Sprintf(`{"status":"%s","environment":"%s"}`, data, dockerInfo.ENVIRONMENT))
	req, err := http.NewRequest("POST", dockerInfo.CALLBACK, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		ctx.WithError(err)
	}

	defer resp.Body.Close()
	ctx.Warn(resp.Status)
}

// BuildDockerContainer executes a docker build command and assigns it a random tag
func BuildDockerContainer(ctx log.Interface, com, quit chan string, buildServerPath string) {
	// First we will seed the random number generator
	rand.Seed(time.Now().UnixNano())
	randomTag := strconv.Itoa(rand.Intn(1000))
	// TODO: REMOVE TEMP LAYERS
	out, err := exec.Command("docker", "build", "--no-cache=True", "-t", randomTag, buildServerPath).Output()
	ctx.Info(string(out))
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
	ctx := log.WithFields(log.Fields{
		"repo": dockerInfo.REPO_NAME,
		"env":  dockerInfo.ENVIRONMENT,
	})

	// Create a new Dockerfile template parses template definition
	docker_tmpl, err := template.ParseFiles("templates/dockerfile.tmplt")
	if err != nil {
		ctx.WithError(err)
	}

	err_return = err

	f, err := os.Create(buildServerPath + "/Dockerfile")
	if err != nil {
		ctx.WithError(err)
	}
	err_return = err
	defer f.Close()

	//Apply the Dockerfile template to the docker info from the request
	err = docker_tmpl.Execute(f, dockerInfo)
	err_return = err
	if err != nil {
		ctx.WithError(err)
	}

	return err_return
}

// grabBuiltStaticFiles issues a `docker run` command to the container image
// we created that will transfer built files to specified location
func GrabBuiltStaticFiles(ctx log.Interface, dockerImageID, projectName, buildServerPath string) {
	// Not sure if this is the best way to handle "dynamic strings"
	mountStringSlice := []string{buildServerPath, ":", "/tmp_mount"}
	mountString := strings.Join(mountStringSlice, "")

	transfer := exec.Command("scripts/transfer_files.sh", mountString, dockerImageID, projectName)
	res, err := transfer.CombinedOutput()
	if err != nil {
		ctx.WithError(err)
	}
	ctx.Warn(string(res))
}

func Build(buildServerPath string, dockerInfo DockerInfo) string {
	ctx := log.WithFields(log.Fields{
		"repo": dockerInfo.REPO_NAME,
		"env":  dockerInfo.ENVIRONMENT,
	})
	c1 := make(chan string)
	quit := make(chan string)
	go BuildDockerContainer(ctx, c1, quit, buildServerPath)
	updateApiStatus(dockerInfo, "building")

	// Looping until we get notification on the channel c1 that the build has finished
	for {
		select {
		case <-quit:
			ctx.Info("There was an error building the docker container")
			updateApiStatus(dockerInfo, "failed")
			return "fail"
		case buildTag := <-c1:
			ctx.Info("Container built...transfering built files...")
			GrabBuiltStaticFiles(ctx, buildTag, dockerInfo.REPO_NAME, buildServerPath)
			if Config.ENV != "test" {
				rsyncProject(ctx, buildServerPath+"/public/*", Config.DEPLOYROOTPATH+dockerInfo.PATH)
			}
			updateApiStatus(dockerInfo, "success")
			return "success"
		}
	}
}

func rsyncProject(ctx log.Interface, buildServerPath, remoteLoc string) {
	rsyncCommand := exec.Command("scripts/rsync_project.sh", buildServerPath, remoteLoc)
	res, err := rsyncCommand.CombinedOutput()
	if err != nil {
		ctx.WithError(err)
	}
	ctx.Info(string(res))
}

func createTempSSHKey(dockerInfo DockerInfo, buildServerPath string) error {
	var err_return error
	ctx := log.WithFields(log.Fields{
		"repo": dockerInfo.REPO_NAME,
		"env":  dockerInfo.ENVIRONMENT,
	})

	d1 := []byte(dockerInfo.DEPLOY_KEY)
	err := ioutil.WriteFile(buildServerPath+"/id_rsa", d1, 0644)
	if err != nil {
		ctx.WithError(err)
	}
	err_return = err

	return err_return
}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	log.SetHandler(cli.New(os.Stdout))

	ctx := log.WithFields(log.Fields{
		"repo": dockerInfo.REPO_NAME,
		"env":  dockerInfo.ENVIRONMENT,
	})
	ctx.Info(fmt.Sprintf("Started building %s", dockerInfo.REPO_NAME))

	// Let's hold a reference to the project's path to build on
	buildServerPath := Config.BUILDLOCATION + "/" + dockerInfo.PATH

	err := os.MkdirAll(buildServerPath, 0770)
	err = createTempSSHKey(dockerInfo, buildServerPath)
	err = GenerateDockerFile(dockerInfo, buildServerPath)

	if err != nil {
		ctx.WithError(err)
		r.JSON(500, map[string]interface{}{"detail": "failed to build"})
		updateApiStatus(dockerInfo, "failed")
	}

	ctx.Info("Dockerfile generated successfully...building container...")

	// Need to pass in more informabout about the location of
	go Build(buildServerPath, dockerInfo)

	r.JSON(201, nil)
}
