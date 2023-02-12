package lib

import "gov/config"

type Class struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Tenant      string   `yaml:"tenant"`
	Repo        string   `yaml:"repo"`
	Arch        []string `yaml:"arch"`
	config      config.Class
}
