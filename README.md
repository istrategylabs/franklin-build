# Franklin Build

## Installation

1. Since this project is very lightweight and requires building docker images,
   we are NOT currently run it using docker to avoid "Docker in Docker" (DinD) 
   for the moment. This will likely change in the future as the need arises. 
1. Install python 2.7
1. Install [docker toolbox](https://www.docker.com/toolbox)
1. `pip install Flask==0.10.1`

## Running
1. `python build-api/api.py`
1. Make a POST request to `localhost:5000/build` with a body similar to: 

    ```
    {
      "repo_name": "istrategylabs/franklin-api",
      "git_hash": "b6046c5bef74edfc1cbf35f97f62cebdadf6946a",
      "path": "istrategylabs/franklin-api"
    }
    ```
1. A successful response will look like:

    ```
    { 
      "deployed": true, 
      "error": "", 
      "url": "http://www.google.com/" 
    }
    ```

1. A failed response will look like:

    ```
    { 
      "deployed": false, 
      "error": "Something went wrong", 
      "url": "" 
    }
    ```

1. Note: this application creates a `tmp` dir for doing it's work. You may need
   to cleanup/delete this manually at times. It should be in the same location
   as your code.

## TODO

- Include information about Jinja2 compiling and environment variables
