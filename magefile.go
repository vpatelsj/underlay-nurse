// +build mage

// This is a magefile for compiling and releasing underlay-nurse. It will show you the targets
// defined in this magefile
package main

import (
	"fmt"
	"log"
	"golang.org/x/crypto/ssh"
	"github.com/magefile/mage/sh"
    "io/ioutil"
)
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
	return nil
}

func Deploy() error {

	fmt.Println("Removing old Files for vmss0")
	if err := runCmd("40.64.81.159:50000", "sudo rm -Rf /home/azureuser/storage.env && sudo rm -Rf /home/azureuser/underlay-nurse "); err != nil {
		return err
	}
	fmt.Println("Removing Files for vmss3")
	if err := runCmd("40.64.81.159:50003", "sudo rm -Rf /home/azureuser/storage.env && sudo rm -Rf /home/azureuser/underlay-nurse"); err != nil {
		return err
	}

	fmt.Println("Uploading binary to vmss0")
	if err := sh.Run("scp", "-P", "50000","/home/vmpi/go/bin/underlay-nurse","azureuser@40.64.81.159:/home/azureuser"); err != nil {
		return err
	}
	fmt.Println("Uploading to env file to vmss0")
	if err := sh.Run("scp", "-P", "50000","/home/vmpi/storage.env","azureuser@40.64.81.159:/home/azureuser"); err != nil {
		return err
	}
	fmt.Println("Moving Files for vmss0")
	if err := runCmd("40.64.81.159:50000", "sudo mv /home/azureuser/storage.env /tmp && sudo mv /home/azureuser/underlay-nurse /tmp"); err != nil {
		return err
	}
	fmt.Println("Run collect-diag vmss0")
	if err := runCmd("40.64.81.159:50000", "nohup /tmp/underlay-nurse collect-diag > foo.out 2> foo.err < /dev/null &"); err != nil {
		return err
	}

	fmt.Println("Uploading binary to vmss3")
	if err := sh.Run("scp", "-P", "50003","/home/vmpi/go/bin/underlay-nurse","azureuser@40.64.81.159:/home/azureuser"); err != nil {
		return err
	}
	fmt.Println("Uploading to env file to vmss3")
	if err := sh.Run("scp", "-P", "50003","/home/vmpi/storage.env","azureuser@40.64.81.159:/home/azureuser"); err != nil {
		return err
	}
	fmt.Println("Moving Files for vmss3")
	if err := runCmd("40.64.81.159:50003", "sudo mv /home/azureuser/storage.env /tmp && sudo mv /home/azureuser/underlay-nurse /tmp"); err != nil {
		return err
	}
	fmt.Println("Run collect-diag vmss3")
	if err := runCmd("40.64.81.159:50003", "nohup /tmp/underlay-nurse collect-diag > foo.out 2> foo.err < /dev/null &"); err != nil {
		return err
	}
	return nil
}

func runCmd(host, cmd string) error {

	client, session, err := connectToHost("azureuser", host)
	if err != nil {
		return err
	}
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	client.Close()
	return nil
}

func connectToHost(user, host string) (*ssh.Client, *ssh.Session, error) {


	var hostKey ssh.PublicKey
	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile("/home/vmpi/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}