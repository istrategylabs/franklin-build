import os
import shutil
import subprocess

from flask import Flask, jsonify, render_template, request

app = Flask(__name__)

@app.route('/build', methods=['POST'])
def build():
    json = request.get_json()
    # an error is returned at this point if the json is bad
    repo_name = json.get('repo_name', None)
    repo_owner = json.get('repo_owner', None)
    git_hash = json.get('git_hash', None)

    # TODO - rsync final build results to the 'path' location
    path = json.get('path', None)

    if repo_name and repo_owner and git_hash and path:
        tmp_dir = 'tmp/'

        # Create our temporary working directory
        try:
            os.makedirs(tmp_dir)
        except OSError, e:
            if e.errno == 17 and os.path.isdir(tmp_dir):
                # directory already exists. This is expected
                pass
            else:
                raise

        # Create a project specific Dockerfile from our template
        filled_template = render_template('dockerfile.tmplt',
                                          REPO_NAME=repo_name,
                                          REPO_OWNER=repo_owner,
                                          BRANCH='docker',
                                          HASH=git_hash)
        with open(tmp_dir + 'Dockerfile', 'w') as f:
            f.write(filled_template)

        # Spin up the docker container to pull and build the project
        startscript = subprocess.Popen('docker build --no-cache=True --tag="franklin_builder:tmp" .',
                                       cwd='tmp',
                                       stdin=subprocess.PIPE,
                                       shell=True)
        error_returned = startscript.wait()
        
        # TODO either rsync the files on success here, or (better), do it from
        # within the Dockerfile for the project. Either way, check for success
        # here. Can probably pass back a custom error/success above to help
        # with that validation.
        
        # Done with the project. Destroy all of our tmp work
        stopscript = subprocess.Popen('../build-api/scripts/tear_down_project.sh',
                                      cwd='tmp',
                                      stdin=subprocess.PIPE,
                                      shell=True)
        stopscript.wait()
        shutil.rmtree(tmp_dir)

        # Determine our response based on pass/fail
        if not error_returned:
            # TODO we should we return immediately with a "request accepted" 
            # We don't want api calls waiting a while for a response. 
            # We will either want a status endpoint for api to check in on or 
            # do a webhook/callback into api to update the status.
            return jsonify(deployed=True, 
                           url='http://www.google.com/',
                           error='')
    return jsonify(deployed=False, 
                   url='',
                   error='Something went wrong')

if __name__ == "__main__":
    app.debug = True
    app.run('0.0.0.0')
