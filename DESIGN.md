# Application design

Step: List existing plugins in a Jenkins installation

1. cli with option
    jplugins list-installed [--jenkins-home path]
2. loop on files *.[hj]pi to determine plugin name
3. read version in MANIFEST file
4. print out the list

Step: Read public json file and compare

1. Read source ie https://updates.jenkins.io/current/update-center.actual.json
    https://updates.jenkins.io/pluginCount.txt provides the actual count of plugins. can be used to check if load is ok.
2. list installed plugins and loop on each of them
3. if version change, identify the change and add missing dependency (new one)
4. display what changed