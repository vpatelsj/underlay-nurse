// +build mage

// This is a magefile for compiling and releasing underlay-nurse. It will show you the targets
// defined in this magefile
package main

import "github.com/magefile/mage/sh"

// Does something pretty cool.
func Build() error {
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	return sh.Run("go", "install", "./...")
}