//+build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	Default = Build
	Aliases = map[string]interface{}{
		"print-cfg": PrintCfg,
	}
)

// Install dependencies
func Deps() error {
	return sh.Run("go", "mod", "download")
}

// Install dependencies and build the main binary
func Build() error {
	mg.Deps(Deps)

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	return sh.RunWith(map[string]string{"GOBIN": fmt.Sprintf("%s/bin", pwd)}, "go", "install", "./...")
}

// Format all files of the project with gofmt
func Fmt() error {
	return sh.RunV("go", "fmt", "./...")
}

// Vet all files
func Vet() error {
	return sh.RunV("go", "vet", "./...")
}

// Print current config after service started
func PrintCfg() error {
	return sh.RunV("go", "run", "cmd/chessex/main.go", "-cfg", "cfg/chessex.json", "-print-cfg")
}

// Install dependencies and load the lichess database into scylladb
func Load() error {
	mg.Deps(Deps)

	return sh.RunV("go", "run", "cmd/chessex/load.go", "-cfg", "cfg/chessex.json")
}

// Run all tests
func Test() error {
	mg.Deps(Deps)

	return sh.RunV("go", "test", "./...")
}
