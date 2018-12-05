package coremgt

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"jplugins/simplefile"

	"github.com/forj-oss/utils"

	"github.com/forj-oss/go-git"

	"github.com/forj-oss/forjj-modules/trace"
)

// PluginsStatus represents the jenkins update status
type PluginsStatus struct {
	plugins       map[string]*pluginsStatusDetails
	groovies      map[string]*GroovyStatusDetails
	PluginsStatus map[string]*pluginsStatusDetails
	installed     *ElementsType
	ref           *Repository
	repoPath      string
	repoURL       []*url.URL
	useLocal      bool
}

// NewPluginsStatus creates an a plugin update status with a Ref repository
func NewPluginsStatus(installed *ElementsType, ref *Repository) (pluginsCompared *PluginsStatus) {
	pluginsCompared = new(PluginsStatus)
	pluginsCompared.plugins = make(map[string]*pluginsStatusDetails)
	pluginsCompared.groovies = make(map[string]*GroovyStatusDetails)
	pluginsCompared.installed = installed
	pluginsCompared.ref = ref
	pluginsCompared.repoURL = make([]*url.URL, 0, 3)
	return
}

// WriteSimple write list of plugins and groovies in a simple file format.
func (s *PluginsStatus) WriteSimple(file string) (err error) {
	lockFile := simplefile.NewSimpleFile(file, 3)

	for name, plugin := range s.plugins {
		lockFile.Add(1, "plugin", name, plugin.newVersion.String())
	}
	for name := range s.groovies {
		lockFile.Add(1, "groovy", name)
	}

	err = lockFile.WriteSorted(":")
	if err != nil {
		err = fmt.Errorf("Unable to write '%s'. %s", file, err)
	}
	return
}

// SetLocal set the useLocal to true
// When set to true, jplugin do not clone a remote repo URL to store on cache
func (s *PluginsStatus) SetLocal() {
	if s == nil {
		return
	}
	s.useLocal = true
}

// SetFeaturesRepoURL validate and store the repoURL
// It can be executed several times to store multiple URLs
func (s *PluginsStatus) SetFeaturesRepoURL(repoURL string) error {
	if s == nil {
		return nil
	}

	if repoURLObject, err := url.Parse(repoURL); err != nil {
		return fmt.Errorf("Invalid feature repository URL. %s", repoURL)
	} else {
		s.repoURL = append(s.repoURL, repoURLObject)
	}
	return nil
}

// SetFeaturesPath defines where a features repository is located like `jenkins-install-inits`
func (s *PluginsStatus) SetFeaturesPath(repoPath string) error {
	if s == nil {
		return nil
	}

	if p, err := utils.Abs(repoPath); err != nil {
		return fmt.Errorf("Invalid feature repository path. %s", repoPath)
	} else {
		s.repoPath = p
	}
	return nil
}

// NewInstall consider the installed list as newly installed.
// So, the display updates will display it all when we newly install it.
func (s *PluginsStatus) NewInstall() {
	for name, plugin := range s.installed.list["plugin"] {
		s.plugins[name] = plugin.AsNewPluginsStatusDetails(s.installed)
	}
}

// Compare only plugins against repository.
// TODO: Compare groovies
func (s *PluginsStatus) Compare() (_ error) {
	
	elements := s.installed.list[pluginType]
	ref := s.ref
	for name, plugin := range elements {
		refPlugin, found := ref.Get(name)
		if !found {
			s.obsolete(plugin)
			continue
		}

		if curVer, err := plugin.GetVersion(); err != nil {
			gotrace.Error("Invalid manifest version for `%s`", name)
			continue
		} else if latestVer, errLatest := refPlugin.GetVersion(); errLatest != nil {
			gotrace.Error("Invalid latest version for `%s`", name)
			continue
		} else {
			if curVer.Get().LessThan(latestVer.Get()) {
				version, err := plugin.GetVersion()
				if err != nil {
					return err
				}
				s.addPlugin(version, refPlugin)
			}
		}

		for _, dep := range refPlugin.Dependencies {
			if dep.Optionnal {
				return
			}
			if _, found = elements[dep.Name]; !found {

				if p, found := ref.Get(dep.Name); found {
					s.addPlugin(VersionStruct{}, p)
				} else {
					gotrace.Trace("Internal repo error: From '%s', dependency '%s' has not been found.", name, dep.Name)
					continue
				}
			}
		}
	}
	return nil
}

// chooseNewVersion change the default new version of a locked plugin
func (s *PluginsStatus) chooseNewVersion(name, version string) (_ bool) {
	pluginLock, found := s.plugins[name]
	if !found {
		return
	}

	pluginLock.setVersion(version)
	return true
}

// set do add/update a plugin version to the PluginsStatus structure
// The version given is the current version use.
func (s *PluginsStatus) addPlugin(version VersionStruct, pluginRef *RepositoryPlugin) (ret *pluginsStatusDetails) {
	ret, found := s.plugins[pluginRef.Name]

	if found {
		return nil
	}

	ret = newPluginsStatusDetails().initFromRef(version, pluginRef)
	s.plugins[pluginRef.Name] = ret

	return
}

func (s *PluginsStatus) addGroovy(name string, sourcePath string) (ret *GroovyStatusDetails) {
	groovy, found := s.groovies[name]

	if !found {
		groovy = newGroovyStatusDetails(name, sourcePath)
	}
	groovy.defineVersion(!found)
	return groovy
}

func (s *PluginsStatus) obsolete(plugin Element) {
	_, found := s.plugins[plugin.Name()]

	if found {
		return
	}

	s.plugins[plugin.Name()] = newPluginsStatusDetails().initAsObsolete(plugin)
}

// DisplayUpdates show updates as result
func (s *PluginsStatus) DisplayUpdates() (_ bool) {
	if s == nil {
		return
	}

	if len(s.plugins) == 0 && len(s.groovies) == 0 {
		fmt.Print("No plugins or groovies updates detected.")
		return true
	}

	s.PluginsStatus = make(map[string]*pluginsStatusDetails)
	pluginsList, iMaxTitle := s.sortPlugins()

	fmt.Print("\nPlugins:\n==========\n+-- New plugin\n|+- Latest version\nvv\n")

	iCountUpdated := 0
	iCountNew := 0
	for _, title := range pluginsList {
		plugin := s.PluginsStatus[title]
		if plugin == nil {
			continue
		}
		latestTag := " "
		newTag := " "
		if plugin.latest {
			latestTag = "X"
		}
		if old := plugin.oldVersion.String(); old == plugin.newVersion.String() {
			fmt.Printf("%s%s | %-"+strconv.Itoa(iMaxTitle+3)+"s : %s\n", newTag, latestTag, title+" ("+plugin.name+")", old)
		} else {
			iCountUpdated++
			if old == "new" {
				iCountNew++
				newTag = "X"
				old = ""
			}
			fmt.Printf("%s%s | %-"+strconv.Itoa(iMaxTitle+3)+"s : %-10s => %s\n", newTag, latestTag, title+" ("+plugin.name+")", old, plugin.newVersion)
		}

	}

	pluginsList = make([]string, len(s.groovies))
	iCountGroovy := 0
	iMaxTitle = 0
	for _, groovy := range s.groovies {
		pluginsList[iCountGroovy] = groovy.name
		if val := len(groovy.name); val > iMaxTitle {
			iMaxTitle = val
		}
		iCountGroovy++
	}

	sort.Strings(pluginsList)

	fmt.Print("\nGroovies:\n==========\n+-- New Groovy file\n|+- Latest version\nvv\n")

	iCountGroovyUpdated := 0
	iCountGroovyNew := 0
	for _, name := range pluginsList {
		groovy := s.groovies[name]
		latestTag := " "
		newTag := " "
		if old := groovy.oldCommit; old == groovy.newCommit {
			fmt.Printf("%s%s | %-"+strconv.Itoa(iMaxTitle)+"s : %s\n", newTag, latestTag, name, groovy.newCommit)
		} else {
			iCountGroovyUpdated++
			if old == "" {
				iCountGroovyNew++
				old = "new"
				newTag = "X"
			}
			fmt.Printf("%s%s | %-"+strconv.Itoa(iMaxTitle)+"s : %-30s => %s\n", newTag, latestTag, name, old, groovy.newCommit)
		}

	}

	fmt.Printf("\nFound %d/%d plugin(s) updates available. %d are new.\n", iCountUpdated, len(s.plugins), iCountNew)
	fmt.Printf("Found %d/%d groovy(ies) updates available. %d are new.\n", iCountGroovyUpdated, iCountGroovy, iCountGroovyNew)

	return true
}

// ImportInstalled import a list of plugins considered as pre-installed
// It stores the list of plugins and constraints '>=' in the structure, so those plugins are installed at minimum to that list.
func (s *PluginsStatus) ImportInstalled(elements *ElementsType) {

	pluginsData := elements.list[pluginType]
	for name, plugin := range pluginsData {
		if _, found := s.plugins[name]; found {
			gotrace.Warning("plugin '%s' is duplicated. Ignored. FYI, it is better to extract pre-installed version list from a fresh jenkins installation, like with Docker.", name)
			continue
		}

		pluginRef, found := s.ref.Get(name)
		if !found {
			s.plugins[name] = newPluginsStatusDetails().
				initAsObsolete(plugin)
		} else {
			version, _ := plugin.GetVersion()
			s.plugins[name] = newPluginsStatusDetails().
				initFromRef(version, pluginRef).
				setAsPreInstalled().
				addConstraint(">=" + version.String())
		}
	}
}

// CheckElement call split function and ensure a type and a name are given.
func CheckElement(fields []string, split func(string, string, string)) {
	var ftype, fname string
	var fversion string

	switch len(fields) {
	case 1:
		gotrace.Warning("Line format is incorrect. It should be <'plugin'|'feature'>:<plugin Name>[:<version>]")
		return
	case 2:
		ftype = strings.Trim(fields[0], " ")
		fname = strings.Trim(fields[1], " ")
	default:
		ftype = strings.Trim(fields[0], " ")
		fname = strings.Trim(fields[1], " ")
		fversion = strings.Trim(fields[2], " ")
	}

	split(ftype, fname, fversion)
}

// CheckElementLine analyzes a line and split with the split function.
func (s *PluginsStatus) CheckElementLine(line string, split func(string, string, string)) {
	if line == "" || line[0] == '#' {
		return
	}

	fields := strings.Split(line, ":")
	CheckElement(fields, split)
}

func (s *PluginsStatus) CheckFeature(name string) (_ error) {
	if s == nil {
		return
	}
	if !s.useLocal {
		return fmt.Errorf("Git clone of repository not currently implemented. Do git task and use --features-repo-path")
	}

	if err := git.RunInPath(s.repoPath, func() error {
		if git.Do("rev-parse", "--git-dir") != 0 {
			return fmt.Errorf("Not a valid GIT repository")
		}
		return nil
	}); err != nil {
		return fmt.Errorf("Issue with '%s', %s", s.repoPath, err)
	}

	featureFile := path.Join(s.repoPath, name, name+".desc")
	fd, err := os.Open(featureFile)
	if err != nil {
		return fmt.Errorf("Unable to read feature file '%s'. %s", featureFile, err)
	}
	defer fd.Close()

	fileScan := bufio.NewScanner(fd)
	for fileScan.Scan() {
		line := strings.Trim(fileScan.Text(), " \n")
		if gotrace.IsInfoMode() {
			fmt.Printf("== >> %s ==\n", line)
		}
		s.CheckElementLine(line, func(ftype, fname, version string) {
			switch ftype {
			case "groovy":
				err = s.CheckGroovy(path.Join(name, fname), s.repoPath)
			case "plugin":
				err = s.CheckPlugin(fname, version, nil)
			default:
				gotrace.Warning("feature type '%s' is currently not supported. Ignored.", ftype)
				return
			}
		})
		if err != nil {
			break
		}
	}
	return err
}

func (s *PluginsStatus) CheckGroovy(name, groovyPath string) error {

	groovy, found := s.groovies[name]
	if !found {
		if groovy = s.addGroovy(name, groovyPath); groovy == nil {
			return fmt.Errorf("Unable to add %s several times", name)
		}
		gotrace.Info("New groovy '%s' identified.", name)
		s.groovies[name] = groovy
	}
	return nil
}

func (s *PluginsStatus) CheckPlugin(name, versionConstraints string, parentDependency *pluginsStatusDetails) error {
	refPlugin, found := s.ref.Get(name)
	if !found {
		return fmt.Errorf("Plugin '%s' not found in the public repository", name)
	}

	plugin, found := s.plugins[name]
	if !found {
		if plugin = s.addPlugin(VersionStruct{}, refPlugin); plugin == nil {
			return fmt.Errorf("Unable to add %s several times", name)
		}
		gotrace.Info("New plugin '%s' identified.", name)
	}

	if versionConstraints != "" {
		if parentDependency != nil {
			plugin.setMinimumVersionDep(versionConstraints)
		} else {
			plugin.addConstraint(versionConstraints)
		}
	}

	for _, dep := range refPlugin.Dependencies {
		if dep.Optionnal {
			continue
		}
		s.CheckPlugin(dep.Name, dep.Version, plugin)
	}
	return nil
}

// DefinePluginsVersion will apply latest version of each plugin except if jplugins.lst or *.desc apply a constraints
func (s *PluginsStatus) DefinePluginsVersion() (_ bool) {
	for name, plugin := range s.plugins {
		refPlugin, found := s.ref.Get(name)
		if !found {
			gotrace.Error("Plugin '%s' not found in the public repository. Ignored.", name)
			return
		}
		if foundVersion, latest, err := refPlugin.DetermineVersion(plugin.rules); err != nil {
			gotrace.Error("Unable to find a version for plugin '%s' which respect all rules. %s. Please fix it", name, err)
		} else {
			plugin.setVersion(foundVersion.String())
			if latest {
				plugin.setIsLatest()
			}
			if plugin.oldVersion.String() != foundVersion.String() {
				gotrace.Trace("%s : %s => %s\n", name, plugin.oldVersion, foundVersion)
			} else {
				gotrace.Trace("%s : %s => No update\n", name, plugin.oldVersion)
			}

		}
	}

	return true
}

func (s *PluginsStatus) CheckMinDep() (_ bool) {
	// Detect dependency request issue
	for name, plugin := range s.plugins {
		refplugin, _ := s.ref.Get(name)

		for _, dep := range refplugin.Dependencies {
			depPlugin := s.plugins[dep.Name]
			depPlugin.checkMinimumVersionDep(dep.Version, plugin)
		}
	}

	// display dependency issue
	for name, plugin := range s.plugins {
		minVersion := plugin.minDepVersion.Get()
		if minVersion == nil {
			gotrace.Trace("No dependency identified for %s", name)
			continue
		}
		curVersion := plugin.newVersion.Get()
		gotrace.Trace("%s: Testing selected version (%s) with dependencies constraints '%s'", name, curVersion, minVersion)
		if curVersion.LessThan(minVersion) {
			refplugin, _ := s.ref.Get(name)
			gotrace.Error("%s: Your settings has fixed %s as latest acceptable version. (latest is %s) But %sso, the dependencies requires an higher version %s. "+
				"You have to fix it. Update %s to newer version or downgrade the dependency to lower version to accept %s:%s",
				name, plugin.newVersion, refplugin.Version, plugin.minDepName, plugin.minDepVersion, name, name, plugin.newVersion)
			return
		}
	}
	return true
}

func (s *PluginsStatus) sortPlugins() (pluginsList []string, maxTitle int) {
	pluginsList = make([]string, len(s.plugins))
	iCount := 0
	for pluginName, plugin := range s.plugins {
		if plugin == nil {
			gotrace.Error("Internal issue: Unable to find a valid plugin called '%s'.", pluginName)
			continue
		}
		pluginsList[iCount] = plugin.title
		s.PluginsStatus[plugin.title] = plugin
		if val := len(plugin.title) + len(plugin.name); val > maxTitle {
			maxTitle = val
		}
		iCount++
	}
	sort.Strings(pluginsList)
	return
}

// PluginsLength return the number of plugins candidate for updates
func (s *PluginsStatus) PluginsLength() int {
	return len(s.plugins)
}