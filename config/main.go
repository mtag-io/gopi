package config

import (
	"gopkg.in/yaml.v3"
	"log"
)

func New(rawConfig []byte, rawTpl []byte) *Class {
	this := Class{}

	err := yaml.Unmarshal(rawConfig, &this)
	if err != nil {
		log.Fatalln("Unable to parse configuration file.")
	}

	this.Tpl = string(rawTpl)
	return &this
}
