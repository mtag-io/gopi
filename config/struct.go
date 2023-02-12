package config

type Class struct {
	PkgInfoFile string   `yaml:"pkgInfoFile"`
	IconPath    string   `yaml:"iconPath"`
	ArchList    []string `yaml:"archList"`
	ReadmeFile  string   `yaml:"readmeFile"`
	Tpl         string
}
