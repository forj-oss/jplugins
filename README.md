# Jenkins plugins tool

Jenkins plugins tool is a GO program to manage Jenkins plugins outside Jenkins.

Why do we need to do that?

Jenkins itself has all functions to manage all plugins. (install/update, alerts, ...) 
It offers also some functionnality from API perspective.

So, when we want to update Jenkins and or plugins, we as jenkins to do it.
But this task is done on Jenkins running in production.

In DevOps world, we would like to test this before and even prepare a package with those 
changes so that we just need to shutdown jenkins and restart on the new version.

With Docker, we have tools to prepare new jenkins images with all updates wanted.
So, from Operation point of view, updating Jenkins could be as simple as doing:

```bash
docker pull jenkins:<V2>
docker rm -f jenkins-master # or do a safe shutdown then remove
docker run -it -d ... jenkins:<V2>
```

With Jenkins docker image, we can specify a collection of plugins (plugins.txt)
to install them. ut as soon as installed, they gets never updated by this mechanism

As well, Jenkins and plugins configuration page are not managed (set and maintain)

That's where jenkins-install-inits comes in the view:

The project add notion of Jenkins features, which enhance installation plugin by
installing plugins and groovy files (install.groovy.d), so that when we install
plugin, we can configure and maintain those plugins or Jenkins configuration from 
code.

But there is some limitations in it:

- dependencies are not installed to latest version by default. 
- features are not versionned
- dependencies installation is slow.

So, this repository has been created to replace the shell script to install features/plugins
by a new GO program called `jplugins`

Some of needs to cover:

1. Easy installation

    Instead of shell script, build it in GO to get a linux binary, without library dependencies.
    This binary can be installed easily by simply download and execute.

2. List existing plugins in a Jenkins installation

    `jplugins` should extract existing plugins installation, with versions used to compare with public jenkins-ci.org

3. Read public repos compare and report plugins updates

    `jplugins` should read jenkins-ci.org and compare jenkins installation to identify plugins to update.

4. Lock versions

    `jplugins` accept to build a list of plugins to install and lock versions in a lock file which will be used
    to download and install plugins/features

5. Download and install from locked versions

    During docker build to build the final Jenkins image to start in production, we install jplugins and read the 
    lock file to download and install files in Jenkins, so that Jenkins has all plugins already up to date but 
    respecting any rules declared in the code.

## Development phases

Identified in 4 steps to answer needs described above.

We are in step 4 - Lock version for docker build.

You can contribute to the project. Follow the usual Issues/Pull Request mechanism

## Build and develop

Requirements:

- docker 1.9 or higher

To build the project, do the following:

```bash
mkdir ~/go/src
cd ~/go/src
git clone https://github.com/forj-oss/jplugins.git
cd jplugins
source build-env.sh    # Loading project build environment for GO.
create-go-build-env.sh # Building the GO builder
go build
```

NOTE1: GO is not required as GO is running from docker.
NOTE2: To use docker go, we load a build environment `source build_env.sh` or `build-env` (alias) to be able to run usual go command with docker.

## How to use it

- How to install it?

    ```bash
    wget https://github.com/forj-oss/jplugins/releases/download/0.0.1/jplugins
    chmod +x ./jplugins
    ```

- How to list plugins?

    ```bash
    $ jplugins list-installed
    [...]
    83 plugins
    ```

    You can use --jenkins-home if your jenkins installation is not in the default `/var/jenkins_home`

- How to check jenkins updates?

    ```bash
    $ jplugins check-updates
    [...]
    bouncycastle API Plugin                             : 2.16.0     => 2.16.3
    inheritance-plugin                                  : 1.5.3      => 2.0.0
    vSphere Plugin                                      : 1.1.11     => 2.17

    Found 45 plugin(s) updates available.
    1480 plugins loaded.
    83 plugins installed.
    ```

Forj team