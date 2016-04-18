# Core Fabric imports
from fabric.api import env, require, run, task, local, roles
from fabric.contrib.project import rsync_project
from fabric.utils import abort

# Third-party app imports
import unipath


PROJECT_DIR = unipath.Path(__file__).ancestor(2)
GIT_ORIGIN = 'git@github.com'


#
# Pulling
#

@task
@roles('build')
def pull():
    """ git pull on all the repositories """
    require('roledefs', provided_by=['production'])
    run("cd %(base)s; git pull %(remote)s %(branch)s; git fetch;" % env)

@task
@roles('build')
def restart_service():
    """ Restarts the builder service """
    require('roledefs', provided_by=['production'])
    run("sudo service franklin-build restart")

@task
@roles('build')
def get_go_deps():
    """ Gets all the dependencies needed to run the software. """
    require('roledefs', provided_by=['production'])
    run("cd %(base)s; go get" % env)

@task
@roles('build')
def deploy():
    """ Deploys builder """
    require('roledefs', provided_by=['production'])
    pull()
    get_go_deps()
    restart_service()

@task
def production():
    """ use production settings """
    env.environment = 'production'
    env.repo = ('origin', 'master')
    env.remote, env.branch = env.repo
    env.base = '/home/franklin//src/github.com/istrategylabs/franklin-build'
    env.user = 'franklin'
    env.roledefs = {
        'build': ['52.20.0.42']
    }
    env.dev_mode = False
