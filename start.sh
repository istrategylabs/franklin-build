#!/bin/bash

cd build-api && gunicorn -w 1 api:app
