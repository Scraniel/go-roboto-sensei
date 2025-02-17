//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// Builds the bot and places the binary in bin/
func Build() error {
	mg.Deps(InstallDeps)
	fmt.Println("Building...")
	return execute("go", "build", "-o", "bin/roboto-sensei", ".")
}

// Builds and tags docker image
func Docker() error {
	fmt.Println("Building docker image...")
	return execute("docker", "build", "-t", "roboto-sensei", ".")
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	// Example:
	// cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
	//return cmd.Run()

	return nil
}

func execute(command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Delete bin/
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("bin/")
}
