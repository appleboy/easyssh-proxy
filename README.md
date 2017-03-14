# easyssh-proxy

[![GoDoc](https://godoc.org/github.com/appleboy/easyssh-proxy?status.svg)](https://godoc.org/github.com/appleboy/easyssh-proxy) [![Build Status](http://drone.wu-boy.com/api/badges/appleboy/easyssh-proxy/status.svg)](http://drone.wu-boy.com/appleboy/easyssh-proxy) [![codecov](https://codecov.io/gh/appleboy/easyssh-proxy/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/easyssh-proxy) [![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/easyssh-proxy)](https://goreportcard.com/report/github.com/appleboy/easyssh-proxy) [![Sourcegraph](https://sourcegraph.com/github.com/appleboy/easyssh-proxy/-/badge.svg)](https://sourcegraph.com/github.com/appleboy/easyssh-proxy?badge)

easyssh-proxy provides a simple implementation of some SSH protocol features in Go.

## Feature

This project is forked from [easyssh](https://github.com/hypersleep/easyssh) but add some features as the following.

* [x] Support plain text of user private key.
* [x] Support key path of user private key.
* [x] Support Timeout for the TCP connection to establish.
* [x] Support SSH ProxyCommand.

```
     +--------+       +----------+      +-----------+
     | Laptop | <-->  | Jumphost | <--> | FooServer |
     +--------+       +----------+      +-----------+

                         OR

     +--------+       +----------+      +-----------+
     | Laptop | <-->  | Firewall | <--> | FooServer |
     +--------+       +----------+      +-----------+
     192.168.1.5       121.1.2.3         10.10.29.68
```

## Usage:

You can see `ssh`, `scp`, `ProxyCommand` on `examples` folder.

### ssh

See [example/ssh.go](./example/ssh.go)

```go
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
```

### scp

See [example/ssh.go](./example/scp.go)

```go
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
```

### SSH ProxyCommand

See [example/proxy.go](./example/proxy.go)

```go
	ssh := &easyssh.MakeConfig{
		User:    "drone-scp",
		Server:  "localhost",
		Port:    "22",
		KeyPath: "./tests/.ssh/id_rsa",
		Proxy: easyssh.DefaultConfig{
			User:    "drone-scp",
			Server:  "localhost",
			Port:    "22",
			KeyPath: "./tests/.ssh/id_rsa",
		},
	}
```
