package lib

import "gov/config"

func New(cfg *config.Class) *Class {
	return &Class{
		config: *cfg,
	}
}
