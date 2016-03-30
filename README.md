# Franklin Build

![franklin](https://s-media-cache-ak0.pinimg.com/236x/d9/f9/97/d9f997346e9e651f152ad98f3ffde330.jpg)

## Contributing

Two things to note about contributing to this project:

1. Follow common Go best-practices as presented in the fantastic [Effective Go](https://golang.org/doc/effective_go.html) document.
1. Use [gofmt](https://golang.org/cmd/gofmt/) before committing any code


## Getting Started

Since this project is very lightweight and requires building docker images, we are NOT currently run it using docker to avoid "Docker in Docker" (DinD) for the moment. This will likely change in the future as the need arises.

* Install [docker toolbox](https://www.docker.com/toolbox)

* Install [Go](http://golang.org/doc/install.html) and set up your [GOPATH](https://golang.org/doc/code.html#GOPATH)

To install, run go get in your go path:
	`go get github.com/istrategylabs/franklin-build`

This will install all packages and dependencies

To install and compile manually use the following steps:

1. `git clone https://github.com/istrategylabs/franklin-build/` 
1. `go get`
1. `cd franklin-build`
1. `make build`

## Environment Variables
The following environment variables will need to be set:

	BUILD_LOCATION
	ENV
	DEPLOY_ROOT_FOLDER

1. Set the BUILD_LOCATION environment variable (location where project will build to): `export BUILD_LOCATION=<location>`
2. Set the ENV environment variable (production or test): `export ENV=test`
4. Set the DEPLOY_ROOT_FOLDER variable to the folder where successfully build projects will be rsync'd to. Nginx in `franklin` will have the same exact setting. `export DEPLOY_ROOT_FOLDER=/var/www/franklin/`

## Running
1. `make run`
1. Make a POST request to `localhost:3000/build` with a body similar to what is found in test/sample_data.json
1. You can run the test suite by running `make test`
