# Franklin Build

![franklin](https://s-media-cache-ak0.pinimg.com/236x/d9/f9/97/d9f997346e9e651f152ad98f3ffde330.jpg)

**Please use [gofmt](https://golang.org/cmd/gofmt/) before commiting any code**

## Installation

1. Since this project is very lightweight and requires building docker images,
   we are NOT currently run it using docker to avoid "Docker in Docker" (DinD) 
   for the moment. This will likely change in the future as the need arises. 
1. Install python 3.5
1. Install [docker toolbox](https://www.docker.com/toolbox)
1. `pip install -r requirements.txt`

## Running

1. Set the build_directory environment variable (location where project will build to): `export BUILD_LOCATION=<location>`
1. Set the ENV environment variable (production or test): `export REMOTE_LOCATION=test`
1. `make run`
1. Make a POST request to `localhost:5000/build` with a body similar to what is found in build_api/test/sample_data.json
1. You can run the test suite by running `make test`
