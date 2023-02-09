package pkg

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

const PKG_INFO = "pkg.info"

func (that *Class) PromptPkg() {

	var err error

	fmt.Println("GO pkg.info initializer:")
	that.Name = prompt("Project name(required):", getValidator("empty"))
	that.Version = prompt("Project version (required & semver )", getValidator("semver"))
	that.Description = prompt("Description of the project (Enter for blank)", getValidator("none"))
	that.Tenant = prompt("Tenant to which the project belongs to (required)", getValidator("empty"))
	that.Repo = prompt("Repository url of the project (Enter for blank)", getValidator("none"))
	res := prompt("Architectures list on which the project should be build (Enter for local only)", getValidator("none"))
	that.Arch, err = archValid(res)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (that *Class) CreatePkg() {
	content, err := yaml.Marshal(that)
	if err != nil {
		log.Fatalf("Unable to stringify the %s`s file content", PKG_INFO)
	}
	err = os.WriteFile(PKG_INFO, content, 777)
	if err != nil {
		log.Fatalf("Unable to write the %s file.", PKG_INFO)
	}
}

func (that *Class) GetPackage(root string) {
	if root == "" {
		root, _ = os.Getwd()
	}
	content, err := os.ReadFile(PKG_INFO)
	if err != nil {
		log.Fatalf("Unable to read the %s`s file from %s", PKG_INFO, root)
	}
	err = yaml.Unmarshal(content, that)
}
