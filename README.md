# easyssh-proxy

[![GoDoc](https://godoc.org/github.com/appleboy/easyssh-proxy?status.svg)](https://pkg.go.dev/github.com/appleboy/easyssh-proxy)
[![Lint and Testing](https://github.com/appleboy/easyssh-proxy/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/easyssh-proxy/actions/workflows/testing.yml)
[![codecov](https://codecov.io/gh/appleboy/easyssh-proxy/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/easyssh-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/easyssh-proxy)](https://goreportcard.com/report/github.com/appleboy/easyssh-proxy)
[![Sourcegraph](https://sourcegraph.com/github.com/appleboy/easyssh-proxy/-/badge.svg)](https://sourcegraph.com/github.com/appleboy/easyssh-proxy?badge)

[繁體中文](./README.zh-tw.md)

easyssh-proxy provides a simple implementation of some SSH protocol features in Go.

## Feature

This project is forked from [easyssh](https://github.com/hypersleep/easyssh) but add some features as the following.

- [x] Support plain text of user private key.
- [x] Support key path of user private key.
- [x] Support Timeout for the TCP connection to establish.
- [x] Support SSH ProxyCommand.

```bash
     +--------+       +----------+      +-----------+
     | Laptop | <-->  | Jumphost | <--> | FooServer |
     +--------+       +----------+      +-----------+

                         OR

     +--------+       +----------+      +-----------+
     | Laptop | <-->  | Firewall | <--> | FooServer |
     +--------+       +----------+      +-----------+
     192.168.1.5       121.1.2.3         10.10.29.68
```

## Usage

You can see detailed examples of the `ssh`, `scp`, `Proxy`, and `stream` commands inside the [`examples`](./_examples/) folder.

### MakeConfig

All functionality provided by this package is accessed via methods of the MakeConfig struct.

```go
  ssh := &easyssh.MakeConfig{
    User:    "drone-scp",
    Server:  "localhost",
    KeyPath: "./tests/.ssh/id_rsa",
    Port:    "22",
    Timeout: 60 * time.Second,
  }

  stdout, stderr, done, err := ssh.Run("ls -al", 60*time.Second)
  err = ssh.Scp("/root/source.csv", "/tmp/target.csv")
  stdoutChan, stderrChan, doneChan, errChan, err = ssh.Stream("for i in {1..5}; do echo ${i}; sleep 1; done; exit 2;", 60*time.Second)
```

MakeConfig takes in the following properties:

| property          | description                                                                                                                                    |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| user              | The SSH user to be logged in with                                                                                                              |
| Server            | The IP or hostname pointing of the server                                                                                                      |
| Key               | A string containing the private key to be used when making the connection                                                                      |
| KeyPath           | The path pointing to the SSH key file to be used when making the connection                                                                    |
| Port              | The port to use when connecting to the SSH daemon of the server                                                                                |
| Protocol          | The tcp protocol to be used: `"tcp", "tcp4" "tcp6"`                                                                                            |
| Passphrase        | The Passphrase to unlock the provided SSH key (leave blank if no Passphrase is required)                                                       |
| Password          | The Password to use to login the specified user                                                                                                |
| Timeout           | The length of time to wait before timing out the request                                                                                       |
| Proxy             | An additional set of configuration params that will be used to SSH into an additional server via the server configured in this top-level block |
| Ciphers           | An array of ciphers (e.g. aes256-ctr) to enable for the SSH connection                                                                         |
| KeyExchanges      | An array of key exchanges (e.g. ecdh-sha2-nistp384) to enable for the SSH connection                                                           |
| Fingerprint       | The expected fingerprint to be returned by the SSH server, results in a fingerprint error if they do not match                                 |
| UseInsecureCipher | Enables the use of insecure ciphers and key exchanges that are insecure and can lead to compromise, [see ssh](#ssh)                            |

NOTE: Please view the reference documentation for the most up to date properties of [MakeConfig](https://pkg.go.dev/github.com/appleboy/easyssh-proxy#MakeConfig) and [DefaultConfig](https://pkg.go.dev/github.com/appleboy/easyssh-proxy#DefaultConfig)

### ssh

See [examples/ssh/ssh.go](./_examples/ssh/ssh.go)

```go
package main

import (
  "fmt"
  "time"

  "github.com/appleboy/easyssh-proxy"
)

func main() {
  // Create MakeConfig instance with remote username, server address and path to private key.
  ssh := &easyssh.MakeConfig{
    User:   "appleboy",
    Server: "example.com",
    // Optional key or Password without either we try to contact your agent SOCKET
    // Password: "password",
    // Paste your source content of private key
    // Key: `-----BEGIN RSA PRIVATE KEY-----
    // MIIEpAIBAAKCAQEA4e2D/qPN08pzTac+a8ZmlP1ziJOXk45CynMPtva0rtK/RB26
    // 7XC9wlRna4b3Ln8ew3q1ZcBjXwD4ppbTlmwAfQIaZTGJUgQbdsO9YA==
    // -----END RSA PRIVATE KEY-----
    // `,
    KeyPath: "/Users/username/.ssh/id_rsa",
    Port:    "22",
    Timeout: 60 * time.Second,

    // Parse PrivateKey With Passphrase
    Passphrase: "1234",

    // Optional fingerprint SHA256 verification
    // Get Fingerprint: ssh.FingerprintSHA256(key)
    // Fingerprint: "SHA256:mVPwvezndPv/ARoIadVY98vAC0g+P/5633yTC4d/wXE"

    // Enable the use of insecure ciphers and key exchange methods.
    // This enables the use of the the following insecure ciphers and key exchange methods:
    // - aes128-cbc
    // - aes192-cbc
    // - aes256-cbc
    // - 3des-cbc
    // - diffie-hellman-group-exchange-sha256
    // - diffie-hellman-group-exchange-sha1
    // Those algorithms are insecure and may allow plaintext data to be recovered by an attacker.
    // UseInsecureCipher: true,
  }

  // Call Run method with command you want to run on remote server.
  stdout, stderr, done, err := ssh.Run("ls -al", 60*time.Second)
  // Handle errors
  if err != nil {
    panic("Can't run remote command: " + err.Error())
  } else {
    fmt.Println("don is :", done, "stdout is :", stdout, ";   stderr is :", stderr)
  }
}
```

### scp

See [examples/scp/scp.go](./_examples/scp/scp.go)

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

See [examples/proxy/proxy.go](./_examples/proxy/proxy.go)

```go
  ssh := &easyssh.MakeConfig{
    User:    "drone-scp",
    Server:  "localhost",
    Port:    "22",
    KeyPath: "./tests/.ssh/id_rsa",
    Timeout: 60 * time.Second,
    Proxy: easyssh.DefaultConfig{
      User:    "drone-scp",
      Server:  "localhost",
      Port:    "22",
      KeyPath: "./tests/.ssh/id_rsa",
      Timeout: 60 * time.Second,
    },
  }
```

NOTE: Properties for the Proxy connection are not inherited from the Jumphost. You must explicitly specify them in the DefaultConfig struct.

e.g. A custom `Timeout` length must be specified for both the Jumphost (intermediary server) and the destination server.

### SSH Stream Log

See [examples/stream/stream.go](./_examples/stream/stream.go)

```go
func main() {
  // Create MakeConfig instance with remote username, server address and path to private key.
  ssh := &easyssh.MakeConfig{
    Server:  "localhost",
    User:    "drone-scp",
    KeyPath: "./tests/.ssh/id_rsa",
    Port:    "22",
    Timeout: 60 * time.Second,
  }

  // Call Run method with command you want to run on remote server.
  stdoutChan, stderrChan, doneChan, errChan, err := ssh.Stream("for i in {1..5}; do echo ${i}; sleep 1; done; exit 2;", 60*time.Second)
  // Handle errors
  if err != nil {
    panic("Can't run remote command: " + err.Error())
  } else {
    // read from the output channel until the done signal is passed
    isTimeout := true
  loop:
    for {
      select {
      case isTimeout = <-doneChan:
        break loop
      case outline := <-stdoutChan:
        fmt.Println("out:", outline)
      case errline := <-stderrChan:
        fmt.Println("err:", errline)
      case err = <-errChan:
      }
    }

    // get exit code or command error.
    if err != nil {
      fmt.Println("err: " + err.Error())
    }

    // command time out
    if !isTimeout {
      fmt.Println("Error: command timeout")
    }
  }
}
```
