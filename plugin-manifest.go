package main

type pluginManifest struct {
	Version        string `yaml:"Plugin-Version"`
	Name           string `yaml:"Extension-Name"`
	ShortName      string `yaml:"Short-Name"`
	JenkinsVersion string `yaml:"Jenkins-Version"`
	LongName       string `yaml:"Long-Name"`
	Dependencies   string `yaml:"Plugin-Dependencies"`
	Description    string `yaml:"Specification-Title"`
}
