# Jenkins plugins tool

- [Jenkins plugins tool](#jenkins-plugins-tool)
  - [Why do we need this project to manage jenkins plugins and more](#why-do-we-need-this-project-to-manage-jenkins-plugins-and-more)
  - [What are the main jplugins features compare to Jenkins](#what-are-the-main-jplugins-features-compare-to-jenkins)
  - [Typical jplugins use case](#typical-jplugins-use-case)
    - [Using jplugins in docker context](#using-jplugins-in-docker-context)
      - [What is the typical process](#what-is-the-typical-process)
    - [Outside docker](#outside-docker)
  - [FAQ](#faq)
  - [Build the project](#build-the-project)

Jenkins plugins tool is a GO program to manage Jenkins plugins outside Jenkins in docker context.

## Why do we need this project to manage jenkins plugins and more

Jenkins itself has all functions to manage all plugins. (install/update, alerts, ...)
It offers also some functionnality from API perspective.

So, when we want to update Jenkins and or plugins, we ask jenkins to do it.
But this task is done on Jenkins running in production.

In DevOps world, we would like to test Jenkins updates before moving it to production. So we could build pipelines to check if new plugins version breaks the way you use Jenkins.
Then you prepare a package with those changes (usually a docker image).
So that we just need to shutdown jenkins, and restart it.

In docker context, restarting Jenkins means: remove the container and recreate on the new docker image version.

Using the [official Jenkins docker image](https://github.com/jenkinsci/docker/blob/master/README.md#preinstalling-plugins), you can specify a collection of plugins (plugins.txt) to install.
But the script (`install-plugins.sh`) has a lot of limitations, even if it enhance time to time:

- shell script.
- pinned plugins dependencies are not managed properly.
- manage only plugins (no groovies)
- works only to create a new jenkins instance. No way to control plugins from code perspective

So, this repository has been created to implement a Jenkins As Code mechanism which replace the `install-plugins.sh` to install plugins and more (groovies)

## What are the main jplugins features compare to Jenkins

feature                                                             | Jenkins UI |`install-plugins.sh`| jplugins
--------------------------------------------------------------------|:----------:|:------------------:|:----------:
Install experimental plugins in Jenkins home                        |     X      |        X           |
Easy installation                                                   |    N/A     |        X           |  0.0.1
List installed plugins found from a Jenkins home.                   |     X      |                    |  0.0.1
Check updates of Jenkins against jenkins updates.                   |     X      |                    |  0.0.2
Define a list of plugins to install                                 |            |        X           |  0.0.3
Lock plugin versions to store in git.                               |            |                    |  0.0.3
Update locked plugins from updates                                  |            |                    |  0.0.3
Manage jenkins features (see note 1 for details)                    |            |                    |  0.0.4 ***1**
Update Jenkins plugins in Jenkins Home                              |     X      |                    |  0.0.5 ***2**
Install only locked plugins                                         |            |                    |  0.0.5
Compare lock files and format the output                            |            |                    |  0.0.7
Check updates from lock or list of plugins                          |            |                    |  0.0.8
Check plugins chksum (sha256) or 404                                |     X      |                    |  0.0.8
Retry if download fails                                             |            |        X           |  0.0.8
Pin plugins versions and downgrades parent dependencies of needed.  |            |                    |  0.0.8
Use multiple features repositories                                  |            |                    | Future version

Note *1:<br>
Jenkins features is a collection of plugins list and groovies files to:

- declare one or more plugins to be installed or updated
- install/update some groovy files in jenkins.groovy.d to configure jenkins, plugins, install pipelines, ...

Note *2:<br>
From version 0.0.5, updating jenkins is made with the lock file intermediate step. <br>
(`jplugin init lock-file && jplugin install`)

## Typical jplugins use case

### Using jplugins in docker context

`jplugins` has been designed to manage Jenkins and plugins from code perspective, specifically in Docker context.

The idea:

- Manage list of required plugins and features (`jplugins.lst`)
- Freeze a collection of plugins and groovy files versions (`jplugins.lock`)
- Install plugins and groovy files in a docker image
- Start Jenkins from the docker image and configure Jenkins and plugins automatically with input data.

#### What is the typical process

1. create `jplugins.lst` or generate an initial version from an existing jenkins home and save it to your GIT repo.

    ```bash
    jplugins init features
    git add jplugins.lst
    [...]
    ```

2. Freeze versions of plugins/groovies (`jplugins.lock`) and save it to your GIT repo

    ```bash
    jplugins init lock-file
    git add lock-file
    [...]
    ```

3. Create the Jenkins image reference with plugins/groovies installed.

    Create a `Dockerfile`:

    ```Dockerfile
    FROM jenkins/jenkins

    ARG JENKINS_INSTALL_INITS_URL=https://github.com/forj-oss/jenkins-install-inits

    RUN curl https://github.com/forj-oss/jplugins/releases/download/latest/jplugins -O /usr/local/bin/jplugins && \
        chmod +x ./jplugins

    COPY jplugins.lock $JENKINS_DATA_REF

    RUN git clone $JENKINS_INSTALL_INITS_URL /tmp/jenkins-install-inits && \
        /usr/local/bin/jplugins install --jenkins-home $JENKINS_DATA_REF --lock-file $JENKINS_DATA_REF/jplugins.lock --features-repo-path=/tmp/jenkins-install-inits
    ```

    And build the image:

    ```bash
    $ docker build . -t myjenkins
    [...]
    Step 7/9 : RUN git clone $JENKINS_INSTALL_INITS_URL /tmp/jenkins-install-inits &&     /usr/local/bin/jplugins install --jenkins-home $JENKINS_DATA_REF --lock-file $JENKINS_DATA_REF/jplugins.lock --features-repo-path=/tmp/jenkins-install-inits
    INFO: 1/2 Loading repositories... update-center.actual.json
    INFO: 2/2 Loading repositories... plugin-versions.json
    INFO: Repositories loaded.
    WARNING ! jplugins/coremgt.(*Plugin).ChainElement: The plugin 'artifactory' has a dependent plugin 'perforce' not found in the public repository. Ignored.
    WARNING ! jplugins/coremgt.(*Plugin).ChainElement: The plugin 'ghprb' has a dependent plugin 'build-flow-plugin' not found in the public repository. Ignored.
    WARNING ! jplugins/coremgt.(*Plugin).ChainElement: The plugin 'monitoring' has a dependent plugin 'aws-cloudwatch-library' not found in the public repository. Ignored.
    - plugin:Office-365-Connector               ...  installed - 4.5       sha256:mzCtSXYqvxQYtKVqg4SEShgEwpPgVwnSeVeVaPeDxI8=
    - plugin:ace-editor                         ...  installed - 1.1       sha256:q8lwKIk8inFYGl9VnqSOjh8aZRZPrultq/7Z6V6autI=
    - plugin:analysis-core                      ...  installed - 1.95      sha256:puY3Y0RLmk1lU9e3Tg0Aqyrafw416So9IS7z50eOijc=
    [...]
    INFO: Total 125 installed: 111 plugins, 14 groovies. 0 obsoleted. 0 error(s) found.
    ```

4. Run Jenkins from your image

    `<JENKINS_CONF1>=<value>` is an example. Supported Jenkins configuration keypair depends on groovies installed.
    <br>Check [jenkins-install-inits repository](https://github.com/forj-oss/jenkins-install-inits) for details

    ```bash
    docker run -e <JENKINS_CONF1>=<value> [...] myjenkins
    ```

### Outside docker

You can also configure an existing Jenkins home on a native server with jplugins.
In this case, you have 2 options:

1. create `jplugins.lst` and `jplugins.lock` file, then install
2. install directly from `jplugins.lst`

In the first case, you will typipcally run the following:

```bash
# create my jplugins.lst
vim jplugins.lst
# lock versions
jplugins init lock-file
# Save them in your GIT repo
git add jplugins.lst jplugins.lock
# Install on it on your jenkins server
scp jplugins.lock myserver:/var/jenkins_home
ssh myserver:jplugins install --lock-file /var/jenkins_home/jplugins.lock
```

in the second case, you will typically run the following with version 0.0.10.
Before this version (from 0.0.5), you must create the `jplugins.lock` step like the first case.

```bash
# create my jplugins.lst
vim jplugins.lst
# Save it in your GIT repo
git add jplugins.lst
# Install on it on your jenkins server
scp jplugins.lst myserver:/var/jenkins_home
ssh myserver jplugins install --features-file /var/jenkins_home/jplugins.lst
```

## FAQ

- How to install it?

    ```bash
    wget https://github.com/forj-oss/jplugins/releases/download/latest/jplugins
    chmod +x ./jplugins
    ```

- How to list Jenkins home installed plugins?

    ```bash
    $ jplugins list-installed
    [...]
    83 plugins
    ```

    You can use --jenkins-home or $JENKINS_HOME if your jenkins installation is not in the default `/var/jenkins_home`
- How to check jenkins updates against Jenkins home?

    ```bash
    $ jplugins check-updates --use-jenkins-home
    [...]
    bouncycastle API Plugin                             : 2.16.0     => 2.16.3
    inheritance-plugin                                  : 1.5.3      => 2.0.0
    vSphere Plugin                                      : 1.1.11     => 2.17

    Found 45 plugin(s) updates available.
    1480 plugins loaded.
    83 plugins installed.
    ```

    jplugins default Jenkins home is /var/jenkins_home.
    You can update it with --jenkins-home or $JENKINS_HOME

- How to create the list of core plugins (`jplugins.lst`) from a current Jenkins home?

    ```bash
    jplugins init features
    ```

    This command creates a `jplugins.lst` locally from the default jenkins home, /var/jenkins_home.
    If you want to refer to a different Jenkins Home, set $JENKINS_HOME or use --jenkins-home.

    `jplugins.lst` is a source file for `jplugins` which identify plugins and features to install to Jenkins.
    Usually, this file must be controlled by GIT.

- How to lock versions to install?

    This will create a `jplugins.lock` from which will be used by `jplugins install`

    ```bash
    jplugins init lock-file
    ```

    `jplugins.lock` is a generated source file for `jplugins` which identify plugins and features version to install to Jenkins.
    Usually, this file must be controlled by GIT.

- How to check and export updates list?

    This example uses a lock file which was generated with `jplugins init`

    ```bash
    jplugins check-updates --use-lock-file --export-result
    ```

    This creates an `updates.json`.

    If you want a custom export, you can use `--export-template`. Templates mechanism is built on top of [GO text/template module](https://golang.org/pkg/text/template/)

    ```bash
    jplugins check-updates --use-lock-file --export-result --export-template=jplugins-msteams.tmpl
    ```

    where jplugins-msteams.tmpl has:

    ```yaml
    [
    {{ range . }}\
        {
            "name": "{{ .Name }}",
            "value": "updated from {{ .OldVersion }} to {{ .NewVersion }}",
        },
    {{ end}}\
    ]
    ```

    Data structure used by the GO template mechanism:

  - Array of plugins identified as updatable:
    - Name: Plugin short name
    - OldVersion: Old plugin version
    - NewVersion: New plugin version
    - Title: plugin title

## Build the project

Requirements:

- docker 1.9 or higher

To build the project, do the following:

```bash
mkdir ~/go/src
cd ~/go/src
git clone https://github.com/forj-oss/jplugins.git
cd jplugins
source build-env.sh    # Loading project build environment for GO.
go build
```

**NOTE1**: GO is not required to be installed on your worksatation as GO is running from docker.

**NOTE2**: To use docker go, we load a build environment by calling `source build_env.sh` or `build-env` (alias).
`go` will become accessible through you PATH

For details on build-env see [build-env repository](https://github.com/forj-oss/build-env)

Forj team
