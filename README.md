# Franklin Build

![franklin](https://s-media-cache-ak0.pinimg.com/236x/d9/f9/97/d9f997346e9e651f152ad98f3ffde330.jpg)

## Installation

1. Since this project is very lightweight and requires building docker images,
   we are NOT currently run it using docker to avoid "Docker in Docker" (DinD) 
   for the moment. This will likely change in the future as the need arises. 
1. Install python 3.5
1. Install [docker toolbox](https://www.docker.com/toolbox)
1. `pip install -r requirements.txt`

## Running
1. `python build-api/api.py`
1. Make a POST request to `localhost:5000/build` with a body similar to: 

    ```
    {
      "repo_name": "franklin-api",
      "repo_owner": "istrategylabs",
      "git_hash": "b6046c5bef74edfc1cbf35f97f62cebdadf6946a",
      "path": "/home/www/projects/istrategylabs/franklin-api"
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
