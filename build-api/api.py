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

    if repo_name and repo_owner and git_hash:
        tmp_dir = 'tmp/'
        if not os.path.isdir(tmp_dir):
            os.makedirs(tmp_dir)

        filled_template = render_template('dockerfile.tmplt',
                                          REPO_NAME=repo_name,
                                          REPO_OWNER=repo_owner,
                                          BRANCH='docker',
                                          HASH=git_hash)
        with open(tmp_dir + 'Dockerfile', 'w') as f:
            f.write(filled_template)

        shutil.copyfile('build-api/templates/docker-compose.tmplt',
                        'tmp/docker-compose.yml')
        
        # Spin up the docker container to pull and build the project
        startscript = subprocess.Popen('../build-api/scripts/pull_and_build_project.sh',
                                       cwd='tmp',
                                       stdin=subprocess.PIPE,
                                       shell=True)
        error_returned = startscript.wait()
        
        if not error_returned:
            # TODO run any special commands needed against the docker container
            # for the project here.

            #current_dir = subprocess.check_output(['ls', '-l'])
            #docker-compose run web ls /milagro-tequila/client

            # Done deploying the project. Destroy all of our tmp work
            stopscript = subprocess.Popen('../build-api/scripts/tear_down_project.sh',
                                       cwd='tmp',
                                       stdin=subprocess.PIPE,
                                       shell=True)
            stopscript.wait()

            # TODO should we return immediately with a "request accepted"? We
            # probably don't want api waiting a while for a response. If we do
            # that, we'll probably want some sort of status endpoint for api to
            # check in on or do a webhook/callback into api to update the status.
            return jsonify(deployed=True, 
                           url='http://www.google.com/',
                           error='')
    return jsonify(deployed=False, 
                   url='',
                   error='Something went wrong')

if __name__ == "__main__":
    app.debug = True
    app.run('0.0.0.0')
