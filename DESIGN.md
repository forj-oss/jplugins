# Application design

Step: List existing plugins in a Jenkins installation

1. cli with option
    jplugins list-installed [--jenkins-home path]
2. loop on files *.[hj]pi to determine plugin name
3. read version in MANIFEST file
4. print out the list