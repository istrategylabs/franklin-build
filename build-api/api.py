import os
import subprocess

from flask import Flask, jsonify, render_template, request

app = Flask(__name__)

@app.route('/build', methods=['POST'])
def build():
    json = request.get_json()
    # an error is returned at this point if the json is bad
    repo_name = json.get('repo_name', None)
    git_hash = json.get('git_hash', None)
    path = json.get('path', None)
    if repo_name and git_hash and path:
        tmp_dir = 'tmp/'
        # TODO - pull this out and encapsulate
        # Also, get rid of os commands. Use Popen from subprocess
        if not os.path.isdir(tmp_dir + path):
            os.makedirs(tmp_dir + path)
        filled_template = render_template('dockerfile.tmplt',
                                          REPO_NAME=repo_name,
                                          BRANCH='master')
        with open(tmp_dir + 'Dockerfile', 'w') as f:
            f.write(filled_template)
        # TODO - copt the template for docker-compose.yml into tmp
        # TODO execute the `.sh` script to build the app we have in that dir
        folder = subprocess.call('cd tmp', shell=True) # can't do this. Either
        # use `os.` or use some smarter `Popen` command, telling it where to do
        # it's thing. We're actually going to do this in a shell script
        current_dir = subprocess.check_output(['ls', '-l'])


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
