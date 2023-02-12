package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var isSemver = regexp.MustCompile("^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$")

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

func promptConfirm(label string) bool {
	var s string
	var err error
	r := bufio.NewReader(os.Stdin)

	_, err = fmt.Fprint(os.Stderr, label)
	s, err = r.ReadString('\n')
	if err != nil {
		log.Fatalln("Unable to read/write from/to console.")
	}
	st := strings.TrimSpace(s)
	return st == "y" || st == "yes"
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
			return isSemver.MatchString(strings.TrimSpace(st))
		},
	}
	return v[name]
}

func archValid(st string, archList []string) ([]string, error) {
	if len(strings.TrimSpace(st)) == 0 {
		panic("Empty arch list")
	}

	var lst []string

	al := strings.Split(st, ",")
	for _, a := range al {
		tmp := strings.TrimSpace(a)
		if len(tmp) > 0 && contains(archList, tmp) {
			lst = append(lst, tmp)
		} else {
			fmt.Printf("invalid architecture specification: %s. It wioll be ignored", tmp)

		}
	}
	if len(lst) == 0 {
		fmt.Println("No build architecture specified. Assuming local platform.")
	}
	return lst, nil
}
