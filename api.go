package main

import (
	"bytes"
	"fmt"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"
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
	BUILDLOCATION         string // Location of built files on build server
	DEPLOYROOTPATH        string // The root of where we want to rsync files to
	ENV                   string // Environment of running Franklin-build instance
	AWS_ACCESS_KEY_ID     string // Amazon Creds for uploading files to S3
	AWS_SECRET_ACCESS_KEY string // Complimenting secret for above
	AWS_BUCKET            string // Name of bucket on S3
}

var Config config

// Investigate a better way to do this
func init() {
	log.SetHandler(cli.New(os.Stdout))
	log.SetLevel(log.DebugLevel)
	log.Infof("Initializing Application")

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
	Config.AWS_BUCKET = os.Getenv("AWS_BUCKET")
	if Config.AWS_BUCKET == "" {
		log.Warn("Missing environment variable AWS_BUCKET")
		panic("Missing environment variable AWS_BUCKET")
	}
	Config.AWS_ACCESS_KEY_ID = os.Getenv("AWS_ACCESS_KEY_ID")
	if Config.AWS_ACCESS_KEY_ID == "" {
		log.Warn("Missing environment variable AWS_ACCESS_KEY_ID")
		panic("Missing environment variable AWS_ACCESS_KEY_ID")
	}
	Config.AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
	if Config.AWS_SECRET_ACCESS_KEY == "" {
		log.Warn("Missing environment variable AWS_SECRET_ACCESS_KEY")
		panic("Missing environment variable AWS_SECRET_ACCESS_KEY")
	}
	Config.ENV = os.Getenv("ENV")
}

func logError(ctx log.Interface, err error, function string, msg string, details ...string) {
	ctx.WithField("func", function).WithError(err).Error(msg)
	if len(details) > 0 {
		logDetails(ctx, details[0], 5000)
	}
}

func logDetails(ctx log.Interface, msg string, length ...int) {
	// TODO - Add a feature flag for log verbosity

	// Log length can't be longer than the message
	logLimit := utf8.RuneCountInString(msg)

	// ...or longer than the optional passed length
	if len(length) > 0 && length[0] < logLimit {
		logLimit = length[0]
	} else if logLimit > 500 {
		// Default limit if one isn't passed in
		logLimit = 500
	}
	if logLimit > 0 {
		ctx.Info(msg[0:logLimit] + "...")
	}
}

func updateApiStatus(ctx log.Interface, dockerInfo DockerInfo, data string) {
	if Config.ENV == "test" {
		return
	}

	var payload = fmt.Sprintf(`{"status":"%s","environment":"%s"}`, data, dockerInfo.ENVIRONMENT)
	var jsonStr = []byte(payload)
	req, err := http.NewRequest("POST", dockerInfo.CALLBACK, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		logError(ctx, err, "updateApiStatus", "Failed API callback")
	} else {
		ctx.Info("Updated API - " + payload + " - " + resp.Status)
	}

	defer resp.Body.Close()
}

// BuildDockerContainer executes a docker build command and assigns it a random tag
func BuildDockerContainer(ctx log.Interface, com, quit chan string, buildServerPath string) {
	// First we will seed the random number generator
	rand.Seed(time.Now().UnixNano())
	randomTag := strconv.Itoa(rand.Intn(1000))
	// TODO: REMOVE TEMP LAYERS
	out, err := exec.Command("docker", "build", "--no-cache=True", "-t", randomTag, buildServerPath).Output()

	// TODO - Write this to a file to return build logs
	// if string(out) != "" {
	//  	ctx.Info(string(out))
	//}
	if err != nil {
		logError(ctx, err, "BuildDockerContainer", "docker build failed", string(out))
		quit <- "fail"
	} else {
		ctx.Info("Docker Build succeeded...")
		logDetails(ctx, string(out))
	}
	// Passing along the randomTag associated with the built docker container to the channel 'com'
	com <- randomTag
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())

	// Simple 'health' endpoint for AWS load-balancer health checks and others
	m.Get("/health", SystemHealth)

	// Main processing endpoint
	m.Post("/build", binding.Bind(DockerInfo{}), BuildDockerFile)
	m.Run()
}

func GenerateDockerFile(ctx log.Interface, dockerInfo DockerInfo, buildServerPath string) error {
	var err_return error

	// Create a new Dockerfile template parses template definition
	docker_tmpl, err := template.ParseFiles("templates/dockerfile.tmplt")
	if err != nil {
		logError(ctx, err, "GenerateDockerFile", "ingest Docker template error")
	}

	err_return = err

	f, err := os.Create(buildServerPath + "/Dockerfile")
	if err != nil {
		logError(ctx, err, "GenerateDockerFile", "Create Dockerfile error")
	}
	err_return = err
	defer f.Close()

	//Apply the Dockerfile template to the docker info from the request
	err = docker_tmpl.Execute(f, dockerInfo)
	err_return = err
	if err != nil {
		logError(ctx, err, "GenerateDockerFile", "map info to docker template fail")
	} else {
		ctx.Info("Dockerfile generated successfully...")
	}

	return err_return
}

// grabBuiltStaticFiles issues a `docker run` command to the container image
// we created that will transfer built files to specified location
func GrabBuiltStaticFiles(ctx log.Interface, dockerImageID, projectName, buildServerPath string) {
	// Not sure if this is the best way to handle "dynamic strings"
	mountStringSlice := []string{buildServerPath, ":", "/tmp_mount"}
	mountString := strings.Join(mountStringSlice, "")

	ctx.Info("Transferring built files...")
	transfer := exec.Command("scripts/transfer_files.sh", mountString, dockerImageID, projectName)
	res, err := transfer.CombinedOutput()
	if err != nil {
		logError(ctx, err, "GrabBuiltStatucFiles", "transfer script failure")
	} else {
		ctx.Info("Built files transferring...")
		logDetails(ctx, string(res))
	}
}

func Build(ctx log.Interface, buildServerPath string, dockerInfo DockerInfo) string {
	c1 := make(chan string)
	quit := make(chan string)
	ctx.Info("Building container...")
	go BuildDockerContainer(ctx, c1, quit, buildServerPath)

	// Looping until we get notification on the channel c1 that the build has finished
	for {
		select {
		case <-quit:
			updateApiStatus(ctx, dockerInfo, "failed")
			return "fail"
		case buildTag := <-c1:
			ctx.Info("Container built...")
			GrabBuiltStaticFiles(ctx, buildTag, dockerInfo.REPO_NAME, buildServerPath)
			if Config.ENV != "test" {
				uploadProjectS3(ctx, buildServerPath+"/public/", Config.DEPLOYROOTPATH+dockerInfo.PATH)
			}
			updateApiStatus(ctx, dockerInfo, "success")
			return "success"
		}
	}
}

func uploadProjectS3(ctx log.Interface, localPath, remoteLoc string) {
	ctx.Info("Uploading to S3...")
	walker := make(fileWalk)
	go func() {
		// Gather the files to upload by walking the path recursively.
		if err := filepath.Walk(localPath, walker.Walk); err != nil {
			logError(ctx, err, "rsyncProjectS3", "Walk failed")
		}
		close(walker)
	}()

	// For each file found walking upload it to S3.
	uploader := s3manager.NewUploader(session.New(&aws.Config{Region: aws.String("us-east-1")}))
	for path := range walker {
		rel, err := filepath.Rel(localPath, path)
		if err != nil {
			logError(ctx, err, "rsyncProjectS3", "Failed relative path")
		}
		file, err := os.Open(path)
		if err != nil {
			logError(ctx, err, "rsyncProjectS3", "Open file failed")
			continue
		}
		defer file.Close()
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: &Config.AWS_BUCKET,
			Key:    aws.String(filepath.Join(remoteLoc, rel)),
			Body:   file,
		})
		if err != nil {
			logError(ctx, err, "rsyncProjectS3", "rsync failed")
		}
	}
	ctx.Info("Upload successful...")
}

type fileWalk chan string

func (f fileWalk) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		f <- path
	}
	return nil
}

func rsyncProject(ctx log.Interface, buildServerPath, remoteLoc string) {
	rsyncCommand := exec.Command("scripts/rsync_project.sh", buildServerPath, remoteLoc)
	res, err := rsyncCommand.CombinedOutput()
	if err != nil {
		logError(ctx, err, "rsyncProject", "rsync script failure")
	} else {
		ctx.Info("Rsync successful...")
		logDetails(ctx, string(res))
	}
}

func createTempSSHKey(ctx log.Interface, dockerInfo DockerInfo, buildServerPath string) error {
	var err_return error

	d1 := []byte(dockerInfo.DEPLOY_KEY)
	err := ioutil.WriteFile(buildServerPath+"/id_rsa", d1, 0644)
	if err != nil {
		logError(ctx, err, "createTempSSHKey", "write of id_rsa failed")
	} else {
		ctx.Info("Writing id_rsa to file succeeded...")
	}
	err_return = err

	return err_return
}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	log.SetHandler(cli.New(os.Stdout))
	log.SetLevel(log.DebugLevel)

	ctx := log.WithFields(log.Fields{
		"repo": dockerInfo.REPO_NAME,
		"env":  dockerInfo.ENVIRONMENT,
	})
	ctx.Info("Started building...")

	// Let's hold a reference to the project's path to build on
	buildServerPath := Config.BUILDLOCATION + "/" + dockerInfo.PATH

	err := os.MkdirAll(buildServerPath, 0770)
	err = createTempSSHKey(ctx, dockerInfo, buildServerPath)
	err = GenerateDockerFile(ctx, dockerInfo, buildServerPath)

	if err != nil {
		logError(ctx, err, "BuildDockerFile", "GenerateDockerFile failed")
		r.JSON(500, map[string]interface{}{"detail": "failed to build"})
		updateApiStatus(ctx, dockerInfo, "failed")
	}

	// Need to pass in more informabout about the location of
	go Build(ctx, buildServerPath, dockerInfo)

	r.Status(201)
}

func SystemHealth(p martini.Params, r render.Render) {
	r.JSON(200, map[string]interface{}{"status": "good"})
}
