package main

import (
	"fmt"
	"github.com/appleboy/easyssh-proxy"
	"net/http"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	// Use a HTTP proxy listening on 127.0.0.1:8888 to connect to Proxy/Bastion
	ssh := &easyssh.MakeConfig{
		User:      "drone-scp",
		Server:    "localhost",
		Port:      "22",
		KeyPath:   "./tests/.ssh/id_rsa",
		ProxyInfo: http.ProxyFromEnvironment,
		Proxy: easyssh.DefaultConfig{
			User:    "drone-scp",
			Server:  "localhost",
			Port:    "22",
			KeyPath: "./tests/.ssh/id_rsa",
		},
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
