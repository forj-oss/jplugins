package coremgt

import (
	"fmt"
	"io/ioutil"
	"jplugins/utils"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/forj-oss/forjj-modules/trace"
	forjjutils "github.com/forj-oss/utils"
	yaml "gopkg.in/yaml.v2"
)

const (
	jenkinsHomeGroovyPath  = "init.groovy.d"
	jenkinsHomePluginsPath = "plugins"
)

// JenkinsHome represents the Jenkins home where we install or identify plugins
type JenkinsHome struct {
	homePath string
}

// NewJenkinsHome creates a new JenkinsHome object
func NewJenkinsHome(jenkinsHomePath string) (ret *JenkinsHome) {
	ret = new(JenkinsHome)
	ret.homePath = jenkinsHomePath
	return ret
}

// Install execute an installation of plugins/groovies to the right path.
func (j *JenkinsHome) Install(elements *ElementsType, featureRepoPath string) error {

	j.cleanUp()
	return j.install(elements, featureRepoPath)
}

// IsValid is true if Jenkins home and sub
func (j *JenkinsHome) IsValid() bool {
	if !utils.CheckPath(path.Join(j.homePath)) {
		return false
	}
	if !utils.CheckPath(path.Join(j.homePath, jenkinsHomePluginsPath)) {
		gotrace.Warning("%s was not found in %s.", j.homePath, jenkinsHomePluginsPath)
		return false
	}
	if !utils.CheckPath(path.Join(j.homePath, jenkinsHomeGroovyPath)) {
		gotrace.Warning("%s was not found in %s.", j.homePath, jenkinsHomeGroovyPath)
		return false
	}
	return true
}

// GetPlugins read the list of plugins in Jenkins and load them in a Plugins object
func (j *JenkinsHome) GetPlugins() (elements *ElementsType, _ error) {
	pluginsPath := path.Join(j.homePath, jenkinsHomePluginsPath)

	elements = NewElementsType()
	elements.noRecursiveChain()

	fEntries, err := ioutil.ReadDir(pluginsPath)

	if err != nil {
		return nil, fmt.Errorf("Invalid Jenkins home '%s'. %s", pluginsPath, err)
	}

	var fileRE, manifestRE *regexp.Regexp
	fileREDefine := `^(.*)(\.[jh]pi)$`
	manifestREDefine := `([\w-]*: )(.*)\n`

	if re, err := regexp.Compile(fileREDefine); err != nil {
		gotrace.Error("Internal error. Regex '%s': %s", fileREDefine, err)
		return
	} else {
		fileRE = re

	}
	if re, err := regexp.Compile(manifestREDefine); err != nil {
		gotrace.Error("Internal error. Regex '%s': %s", fileREDefine, err)
		return
	} else {
		manifestRE = re
	}

	for _, fEntry := range fEntries {
		if fEntry.IsDir() {
			continue
		}

		if fileMatch := fileRE.FindAllStringSubmatch(fEntry.Name(), -1); fileMatch != nil {
			pluginFileName := fileMatch[0][0]
			pluginName := fileMatch[0][1]
			pluginExt := fileMatch[0][2]

			if pluginFileName != "" && pluginName == "" {
				gotrace.Error("Invalid file '%s'. Ignored.", pluginFileName)
				continue
			}

			pluginMetafile := path.Join(pluginsPath, pluginName, "META-INF", "MANIFEST.MF")

			tmpExtract := false
			packagePath := path.Join(pluginsPath, pluginName)
			if _, err := os.Stat(pluginMetafile); err != nil && os.IsNotExist(err) {
				if _, s := forjjutils.RunCmdOutput("unzip", "-q", packagePath+pluginExt, "META-INF/MANIFEST.MF", "-d", packagePath); s != 0 {
					gotrace.Error("Unable to extract MANIFEST.MF from plugin package %s", pluginName)
					continue
				}
				tmpExtract = true
			}

			var manifest *Plugin

			if d, err := ioutil.ReadFile(pluginMetafile); err != nil {
				gotrace.Error("Unable to read file '%s'. %s. Ignored", pluginMetafile, err)
				continue
			} else {
				// Remove DOS format if exist
				data := strings.Replace(string(d), "\r", "", -1)
				// and remove new lines ('\n ')
				data = strings.Replace(data, "\n ", "", -1)
				// and escape "
				data = strings.Replace(data, "\"", "\\\"", -1)
				// and embrace value part with "
				data = manifestRE.ReplaceAllString(data, `$1"$2"`+"\n")

				manifest = NewPlugin()
				if err := yaml.Unmarshal([]byte(data), &manifest); err != nil {
					gotrace.Error("Unable to read file '%s' as yaml. %s. Ignored", pluginMetafile, err)
					fmt.Print(data)
					continue
				}
			}
			if tmpExtract {
				os.RemoveAll(packagePath)
				tmpExtract = false
			}
			if _, err = elements.AddElement(manifest); err != nil {
				return nil, err
			}
		}
	}
	return
}

/*******************************************************************************
 ***************************** Internal Functions ******************************
 *******************************************************************************/

// cleanUp remove plugins and groovies before install.
func (j *JenkinsHome) cleanUp() {
	pathsToCheck := []string{
		path.Join(j.homePath, jenkinsHomeGroovyPath),
		path.Join(j.homePath, jenkinsHomePluginsPath),
	}

	filesToCleanUp := make([]*regexp.Regexp, 2)
	filesToCleanUp[0], _ = regexp.Compile(`^.*\.groovy$`)
	filesToCleanUp[1], _ = regexp.Compile(`^.*\.[jh]pi$`)

	for index, pathToCheck := range pathsToCheck {
		if info, err := os.Stat(pathToCheck); os.IsNotExist(err) || !info.IsDir() {
			if err != nil {
				gotrace.Error("Unable to install. Invalid Jenkins home dir. %s", err)
			} else {
				gotrace.Error("Unable to install. Invalid Jenkins home dir. %s is not a directory", pathToCheck, err)
			}
			return
		}

		cleanupDir, _ := ioutil.ReadDir(pathToCheck)
		for _, element := range cleanupDir {
			if fileName := element.Name(); filesToCleanUp[index].MatchString(fileName) {
				if err := os.Remove(path.Join(pathToCheck, fileName)); err != nil {
					gotrace.Error("Cleanup failure. %s", err)
					os.Exit(1)
				}
			}
		}
	}
}

// install plugins and groovies as defined by plugins list.
func (j *JenkinsHome) install(elementsType *ElementsType, featureRepoPath string) error {
	iCountGroovy := 0
	iCountPlugin := 0
	iCountError := 0
	iCountObsolete := 0

	elementsList := make([]string, elementsType.Length())

	iCount := 0
	iMaxName := 0
	iMaxVersion := 0
	for elementType, elements := range elementsType.list {
		for name, element := range elements {
			value := elementType + ":" + name
			elementsList[iCount] = value
			if size := len(value); size > iMaxName {
				iMaxName = size
			}
			if plugin, ok := element.(*Plugin); ok {
				if size := len(plugin.Version); size > iMaxVersion {
					iMaxVersion = size
				}
			}
			iCount++
		}
	}

	nameFormat := "- %-" + strconv.Itoa(iMaxName) + "s ... "
	pluginVersionFormat := "%-" + strconv.Itoa(iMaxVersion) + "s"

	iCount = 0
	sort.Strings(elementsList)

	for _, displayName := range elementsList {
		names := strings.Split(displayName, ":")
		name := names[1]
		elementType := names[0]
		element := elementsType.list[elementType][name]

		switch element.GetType() {
		case groovyType:
			groovy := element.(*Groovy)
			if groovy.CommitID == "" {
				if !gotrace.IsDebugMode() {
					fmt.Printf(nameFormat, displayName)
					fmt.Print("obsolete - removed if found.")
				}
				iCountObsolete++
				continue
			}

			groovyObj := newGroovyStatusDetails(name, featureRepoPath)
			groovyObj.newCommit = groovy.CommitID
			if !gotrace.IsDebugMode() {
				fmt.Printf(nameFormat, displayName)
			}
			if err := groovyObj.installIt(path.Join(j.homePath, jenkinsHomeGroovyPath)); err != nil {
				gotrace.Error("Installation issue. %s. Ignored.", err)
				iCountError++
			} else {
				iCountGroovy++
				iCount++
				if !gotrace.IsDebugMode() {
					fmt.Printf(" installed - %s\n", groovy.CommitID)
				}
			}

		case "plugin":
			plugin := element.(*Plugin)
			if plugin.Version == "" {
				if !gotrace.IsDebugMode() {
					fmt.Printf(nameFormat, displayName)
					fmt.Print("obsolete - removed if found.")
				}
				iCountObsolete++
				continue
			}

			pluginObj := newPluginsStatusDetails()
			pluginObj.setVersion(plugin.Version)
			pluginObj.name = plugin.ExtensionName
			pluginObj.newSha256Version = plugin.checkSumSha256
			if !gotrace.IsDebugMode() {
				fmt.Printf(nameFormat, displayName)
			}
			if err := pluginObj.installIt(path.Join(j.homePath, jenkinsHomePluginsPath)); err != nil {
				gotrace.Error("Installation issue. %s. Ignored.", err)
				iCountError++
			} else {
				iCountPlugin++
				iCount++
				if !gotrace.IsDebugMode() {
					notVerified := ""
					if !pluginObj.checkSumVerified {
						notVerified = " not verified!"
					}
					fmt.Printf(" installed - "+pluginVersionFormat+" sha256:%s%s\n", plugin.Version, pluginObj.newSha256Version, notVerified)
				}
			}
		}
	}

	gotrace.Info("Total %d installed: %d plugins, %d groovies. %d obsoleted. %d error(s) found.\n", iCount, iCountPlugin, iCountGroovy, iCountObsolete, iCountError)
	if iCountError > 0 {
		return fmt.Errorf("%d errors detected", iCountError)
	}

	return nil
}
