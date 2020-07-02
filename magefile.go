// +build mage

// This is a magefile for compiling and releasing underlay-nurse. It will show you the targets
// defined in this magefile
package main

import "github.com/magefile/mage/sh"
import "fmt"
// Does something pretty cool.
func Build() error {
	fmt.Println("Running Go Mod Download")
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	fmt.Println("Running Go install")
	if err := sh.Run("go", "install", "./..."); err != nil {
		return err
	}
	fmt.Println("Uploading to vmss3")
	if err := sh.Run("scp", "-P", "50003","/home/vmpi/go/bin/underlay-nurse","azureuser@40.64.81.159:/home/azureuser"); err != nil {
		return err
	}
	fmt.Println("Uploading to vmss0")
	if err := sh.Run("scp", "-P", "50000","/home/vmpi/go/bin/underlay-nurse","azureuser@40.64.81.159:/home/azureuser"); err != nil {
		return err
	}
	return nil
}