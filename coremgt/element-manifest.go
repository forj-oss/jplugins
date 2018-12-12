package coremgt

// ElementManifest describe details on elements.
// By default, it helps reading the MANIFEST.MF from java as a yaml file, clenaed before.
type ElementManifest struct {
	Version        string `yaml:"Plugin-Version"`
	Name           string `yaml:"Extension-Name"`
	ShortName      string `yaml:"Short-Name"`
	JenkinsVersion string `yaml:"Jenkins-Version"`
	LongName       string `yaml:"Long-Name"`
	Dependencies   string `yaml:"Plugin-Dependencies"`
	Description    string `yaml:"Specification-Title"`
	elementType    string // plugin or groovy
	commitID       string // Commit ID for groovy files
}

func (e *ElementManifest) GetVersion() (ret VersionStruct, err error) {
	err = ret.Set(e.Version)
	return
}

// NewElementManifest return a element manifest with details
func NewElementManifest(typ, name string) (ret *ElementManifest) {
	ret = new(ElementManifest)
	ret.elementType = typ
	ret.Name = name
	return
}
