# Application design

## Planned steps

Step 1: List existing plugins in a Jenkins installation

1. cli with option
    jplugins list-installed [--jenkins-home path]
2. loop on files *.[hj]pi to determine plugin name
3. read version in MANIFEST file
4. print out the list

Step 2: Read public json file and compare

1. Read source ie https://updates.jenkins.io/current/update-center.actual.json
    https://updates.jenkins.io/pluginCount.txt provides the actual count of plugins. can be used to check if load is ok.
2. list installed plugins and loop on each of them
3. if version change, identify the change and add missing dependency (new one)
4. display what changed

Step 3: Read a features.yaml, and pre-installed list pre-install.lst (if exists) and build a features.lock

1. write `jplugins-preinstalled.lst` from an existing installation
2. read `jplugins-preinstalled.lst` and check_updates from this.
3. read `jplugins-preinstalled.lst` and write a simple lock file `jplugins.lock`
4. read `jplugins.lst`, apply rules and write the lock file.

Step 4: Read jplugins.lst, from pre-installed plugins and treat features to add groovy and plugin in the list
Process objective:

    - Apply rules identified in `jplugins.lst` and `*.desc`, and identify latest version to use
    - Display result as list (plugins + groovies)
    - Apply dependencies minimum version required on version identified. 
        It can breaks when an identified version is not compatible with the rule. `jplugins` must report the module with version which requiring a newer version.
        To fix this issue, the `jplugins.lst` can:

    - accept a newer version of the plugin that break the dependency rules.
    - or downgrade the dependency with an older version.

        What it won't do:

    - If the dependency requires an identified newer version, search for an older dependency which accept the identified version (extra step)

1. read `jplugins.lst`, do git clone/update of install-inits, read each desc files requested, add plugins in the list, 
2. list plugins & groovy files

Step 5: Read the lock file, download plugins/groovy files and install them in Jenkins home

1. read `jplugins.lock` and list plugins and groovy files.
2. download plugins is series and install them in Jenkins home (must exist)
3. git clone install-inits and copy files.

Step 6: `jplugins` read lock file and update to latest version and write new version in Lock file

1. read `jplugins.lock` and run updates and display result
2. Test `jplugins.lock` existence calling init command to suggest running update instead.

## Extra steps

step 7: If the dependency requires an identified version newer, search for an older dependency which accept the identified version.

1. enhance the determineVersion with this logic

step 8: Fix feature version on commit ID

1. take feature version as commit ID.

Step 9: Parallelize plugins download to accelerate download (POC)

1. Parallelize the download with GO channel

Step 10: Be able to update partially the Lock file

1. Display update proposal from lock files
2. choose which one to update