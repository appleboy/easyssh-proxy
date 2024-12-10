package main

import (
	"strings"
	"time"

	"github.com/appleboy/easyssh-proxy"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		User:    "appleboy",
		Server:  "example.com",
		KeyPath: "/Users/username/.ssh/id_rsa",
		Port:    "22",
		Timeout: 60 * time.Second,
	}

	fileContents := "Example Text..."
	reader := strings.NewReader(fileContents)

	// Write a file to the remote server using the writeFile command.
	// Second arguement specifies the number of bytes to write to the server from the reader.
	if err := ssh.WriteFile(reader, int64(len(fileContents)), "/home/user/foo.txt"); err != nil {
		panic("Error: failed to write file to client")
	}
}
