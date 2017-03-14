package main

import (
	"fmt"

	"github.com/appleboy/easyssh-proxy"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		User:     "appleboy",
		Server:   "example.com",
		Password: "123qwe",
		Port:     "22",
	}

	// Call Scp method with file you want to upload to remote server.
	// Please make sure the `tmp` floder exists.
	err := ssh.Scp("/root/source.csv", "/tmp/target.csv")

	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		fmt.Println("success")
	}
}
