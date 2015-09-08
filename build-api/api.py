import os, shutil, subprocess, asyncio
import logging
from logging.handlers import RotatingFileHandler

from flask import Flask, jsonify, render_template, request

app = Flask(__name__)
handler = RotatingFileHandler('app.log', maxBytes=10000, backupCount=1)
handler.setLevel(logging.INFO)
app.logger.addHandler(handler)
app.debug = True


def build_docker_container():
    # Spin up the docker container to pull and build the project
    command = 'sudo docker build --no-cache=True --tag="franklin_builder_tmp:tmp" .'
    startscript = subprocess.Popen(
        command,
        cwd='tmp',
        stdin=subprocess.PIPE,
        shell=True
    )
    error_returned = startscript.wait()
    app.logger.info(error_returned)


    # TODO check for success here. Confirm there were no errors during image
    # creation and/or make an external call to confirm the new site is live.

    # Done with the project. Destroy all of our tmp work
    stopscript = subprocess.Popen(
        'scripts/tear_down_project.sh',
        cwd='tmp',
        stdin=subprocess.PIPE,
        shell=True
    )
    stopscript.wait()
    shutil.rmtree('tmp/')

def call_in_background(target, *, loop=None, executor=None):
    """Schedules and starts target callable as a background task

    If not given, *loop* defaults to the current thread's event loop
    If not given, *executor* defaults to the loop's default executor

    Returns the scheduled task.
    """
    if loop is None:
        loop = asyncio.get_event_loop()
    if callable(target):
        return loop.run_in_executor(executor, target)
    raise TypeError("target must be a callable")

@app.route('/build', methods=['POST'])
def build():
    json = request.get_json()
    # an error is returned at this point if the json is bad
    repo_name = json.get('repo_name', None)
    repo_owner = json.get('repo_owner', None)
    git_hash = json.get('git_hash', None)

    path = json.get('path', None)

    if repo_name and repo_owner and git_hash and path:
        tmp_dir = 'tmp/'
        # Create our temporary working directory
        try:
            os.makedirs(tmp_dir)
        except OSError as e:
            if e.errno == 17 and os.path.isdir(tmp_dir):
                # directory already exists. This is expecte
                app.logger.info('Directory already exists')
                pass
            else:
                app.logger.info('temp directory error')
                raise

        # Create a project specific Dockerfile from our template

        filled_template = render_template(
            'dockerfile.tmplt',
            REPO_NAME=repo_name,
            REPO_OWNER=repo_owner,
            BRANCH='docker',
            HASH=git_hash,
            REMOTE_LOC=path
        )

        with open(tmp_dir + 'Dockerfile', 'w') as f:
            f.write(filled_template)

        # We will spawn the docker build in a seperate process
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            threaded_builder = call_in_background(build_docker_container)
        except:
            app.logger.info('Background process setupfailed')
            return jsonify(building=False,
                           error='Background process setupfailed')

        # We will either want a status endpoint for api to check in on or
        # do a webhook/callback into api to update the status.
        app.logger.info('Building')
        return jsonify(building=True, error='')
    app.logger.info('Missing arguments')
    return jsonify(
        building=False,
        error='Please supply all arguments: ' +
              '(repo_name, repo_owner, git_hash, path)'
    )

if __name__ == "__main__":
    app.run(host='127.0.0.1', port=8000)
