package main

import (
	"fmt"

	"github.com/appleboy/easyssh-proxy"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		User:   "appleboy",
		Server: "example.com",
		// Optional key or Password without either we try to contact your agent SOCKET
		//Password: "password",
		Key:     "/.ssh/id_rsa",
		Port:    "22",
		Timeout: 60,
	}

	// Call Run method with command you want to run on remote server.
	stdout, stderr, done, err := ssh.Run("ls -al", 60)
	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		fmt.Println("don is :", done, "stdout is :", stdout, ";   stderr is :", stderr)
	}

}
