package pkg

import (
	"bufio"
	"errors"
	"fmt"
	"gov/version"
	"log"
	"os"
	"strings"
)

var arch = []string{
	"linux_amd64",
	"linux_arm64",
	"darwin_amd64",
	"darwin_arm64",
	"windows",
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func prompt(label string, valid func(st string) bool) string {
	var s string
	var err error
	r := bufio.NewReader(os.Stdin)
	for {
		_, err = fmt.Fprint(os.Stderr, label)
		s, err = r.ReadString('\n')
		if err != nil {
			log.Fatalln("Unable to read/write from/to console.")
		}
		if valid(s) {
			break
		}
	}
	return strings.TrimSpace(s)
}

func getValidator(name string) func(st string) bool {
	v := map[string]func(st string) bool{
		"none": func(st string) bool {
			return true
		},
		// empty - check empty string
		"empty": func(st string) bool {
			return st != ""
		},
		// check string is a valid semver version
		"semver": func(st string) bool {
			err := version.MustParse(st)
			return err != nil
		},
	}
	return v[name]
}

func archValid(st string) ([]string, error) {
	if len(strings.TrimSpace(st)) == 0 {
		panic("Empty arch list")
	}

	var lst []string

	arhList := strings.Split(st, ",")
	for _, a := range arhList {
		tmp := strings.TrimSpace(a)
		if len(tmp) > 0 && contains(arch, tmp) {
			lst = append(lst, tmp)
		} else {
			return nil, errors.New(
				fmt.Sprintf("invalid architecture specification: %s", tmp),
			)
		}
	}
	return lst, nil
}
