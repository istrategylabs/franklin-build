package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"os"
	"os/exec"
)

func buildDockerContainer() {
	command := "sudo docker build --no-cache=True --tags='franklin_builder_tmp:tmp' ."
	if err := exec.Command(command).Run(); err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	m := martini.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Run()
}
