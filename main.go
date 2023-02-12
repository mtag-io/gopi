package main

import (
	_ "embed"
	"flag"
	"gov/config"
	"gov/lib"
	"os"
)

//go:embed config.yaml

var rawConfig []byte

//go:embed readme.tpl
var rawTpl []byte

var initPkg bool
var readMe bool

const usageInitPkg = "Interactively creates a pkg.info file in the current directory"
const usageReadme = "Validates the (if exists) pkg.info file in the current directory"

func init() {
	flag.BoolVar(&initPkg, "init", false, usageInitPkg)
	flag.BoolVar(&initPkg, "i", false, usageInitPkg+" (shorthand)")
	flag.BoolVar(&readMe, "readme", false, usageReadme)
	flag.BoolVar(&readMe, "rm", false, usageReadme+" (shorthand)")
}

func main() {
	flag.Parse()
	root, _ := os.Getwd()
	cfg := config.New(rawConfig, rawTpl)
	gopi := lib.New(cfg)

	if initPkg {
		gopi.PromptPkg(root)
		os.Exit(0)
	}

	if readMe {
		gopi.GetPackage(root)
		gopi.CreateReadme(root, false)
		os.Exit(0)
	}
}
