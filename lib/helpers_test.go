package lib

import (
	"strings"
	"testing"
)

var tArch = []string{
	"linux_amd64",
	"linux_arm64",
	"darwin_amd64",
	"darwin_arm64",
	"windows",
}

func TestHelpers_Contains_include(t *testing.T) {
	tList := []string{"key1", "key2", "key3"}
	if !contains(tList, "key2") {
		t.Fail()
	}
}

func TestHelpers_Contains_exclude(t *testing.T) {
	tList := []string{"key1", "key2", "key3"}
	if contains(tList, "key4") {
		t.Fail()
	}
}

func TestArchValid_ok(t *testing.T) {
	arh, err := archValid("linux_amd64, darwin_arm64", tArch)
	if err != nil {
		t.Fail()
	}
	if strings.Join(arh[:], ", ") != "linux_amd64, darwin_arm64" {
		t.Fail()
	}
}

func TestArchValid_omit(t *testing.T) {
	arh, err := archValid("linux_amd64, darwin_arm64,not-found", tArch)
	if err != nil {
		t.Fail()
	}
	if strings.Join(arh[:], ", ") != "linux_amd64, darwin_arm64" {
		t.Fail()
	}
}

func TestSemverValidator_ok(t *testing.T) {
	sv := getValidator("semver")
	if !sv("1.0.0") {
		t.Fail()
	}
}

func TestSemverValidator_not_ok(t *testing.T) {
	sv := getValidator("semver")
	if sv("1.0.0.wrong") {
		t.Fail()
	}
}
