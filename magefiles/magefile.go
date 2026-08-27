//go:build mage

package main

import (
	"github.com/magefile/mage/sh"
)

// Vet runs go vet on the entire module
func Vet() error {
	return sh.RunV("go", "vet", "./...")
}

// Example runs the example bot from examples/ (it loads .env via godotenv)
func Example() error {
	return sh.RunV("go", "run", "./examples")
}