package lib

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"html/template"
	"log"
	"os"
	"path"
	"strings"
)

func (that *Class) PromptPkg(root string) {

	var err error

	fmt.Println("GO pkg.info initializer:")
	that.Name = prompt("Project name(required):", getValidator("empty"))
	that.Version = prompt("Project version (is required & has to semver compatible): ", getValidator("semver"))
	that.Description = prompt("Description of the project (Enter for blank): ", getValidator("none"))
	that.Tenant = prompt("Tenant to which the project belongs to (required): ", getValidator("empty"))
	that.Repo = prompt("Repository url of the project (Enter for blank): ", getValidator("none"))
	res := prompt("Architectures list on which the project should be build (Enter for local only): ", getValidator("none"))
	that.Arch, err = archValid(res, that.config.ArchList)
	if err != nil {
		log.Fatal(err.Error())
	}
	existingMessage := fmt.Sprintf("A %s file already exists in the %s directory. Overwrite? ( y/yes to confirm): ",
		that.config.PkgInfoFile, root)
	ovr := promptConfirm(existingMessage)
	if ovr {
		that.CreatePkg(root)
	}
}

func (that *Class) checkPkgExists(root string) bool {
	if root == "" {
		root, _ = os.Getwd()
	}
	_, err := os.Stat(path.Join(root, that.config.PkgInfoFile))
	return err == nil
}

func (that *Class) CreatePkg(root string) {
	if root == "" {
		root, _ = os.Getwd()
	}
	raw, err := yaml.Marshal(that)
	if err != nil {
		log.Fatalf("Unable to stringify the %s`s file content", that.config.PkgInfoFile)
	}
	tmp := fmt.Sprintf("# %s pkg.info file\n\n", that.Name) + string(raw)
	err = os.WriteFile(that.config.PkgInfoFile, []byte(tmp), 777)
	if err != nil {
		log.Fatalf("Unable to write the %s file.", that.config.PkgInfoFile)
	}
}

func (that *Class) GetPackage(root string) {
	if root == "" {
		root, _ = os.Getwd()
	}
	content, err := os.ReadFile(that.config.PkgInfoFile)
	if err != nil {
		log.Fatalf("Unable to read the %s`s file from %s.", that.config.PkgInfoFile, root)
	}
	err = yaml.Unmarshal(content, that)
}

func (that *Class) CreateReadme(root string, silent bool) {

	type TplData struct {
		Name        string
		Version     string
		Description string
		Icon        string
	}

	tpl, err := template.New("").Parse(that.config.Tpl)
	if err != nil {
		log.Fatal("Unable to parse the README.md template")
	}
	if root == "" {
		root, _ = os.Getwd()
	}
	var iconPath string
	if !silent {
		msg := fmt.Sprintf("Repo icon file. Defaults to: %s. (Enter for default)", that.config.IconPath)
		iconPath = prompt(msg, getValidator("none"))
	}

	if iconPath == "" {
		iconPath = that.config.IconPath
	}

	tplData := TplData{
		Name:        strings.ToUpper(that.Name),
		Version:     that.Version,
		Description: that.Description,
		Icon:        iconPath,
	}

	pth := path.Join(root, that.config.ReadmeFile)
	fOut, err := os.Create(pth)
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			log.Printf("WARN: Could not close file %s after writing", pth)
		}
	}(fOut)
	if err != nil {
		log.Fatalf("Unable to write %s file in  %s. Check if you have permisssions to do so.",
			that.config.ReadmeFile, root)
	}
	err = tpl.Execute(fOut, tplData)
	if err != nil {
		log.Fatalf("ERROR: while processing README.md template. Reason: %s", err.Error())
	}

}
