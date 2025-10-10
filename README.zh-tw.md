# easyssh-proxy

[![GoDoc](https://godoc.org/github.com/appleboy/easyssh-proxy?status.svg)](https://pkg.go.dev/github.com/appleboy/easyssh-proxy)
[![Lint and Testing](https://github.com/appleboy/easyssh-proxy/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/easyssh-proxy/actions/workflows/testing.yml)
[![codecov](https://codecov.io/gh/appleboy/easyssh-proxy/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/easyssh-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/easyssh-proxy)](https://goreportcard.com/report/github.com/appleboy/easyssh-proxy)
[![Sourcegraph](https://sourcegraph.com/github.com/appleboy/easyssh-proxy/-/badge.svg)](https://sourcegraph.com/github.com/appleboy/easyssh-proxy?badge)

easyssh-proxy 提供了一個用 Go 語言實現的一些 SSH 協議功能的簡單實現。

## 功能

這個項目是從 [easyssh](https://github.com/hypersleep/easyssh) 分叉而來，但添加了一些如下所示的功能。

- [x] 支援用戶私鑰的純文字。
- [x] 支援用戶私鑰的路徑。
- [x] 支援 TCP 連接建立的超時設定。
- [x] 支援 SSH ProxyCommand。

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

## 安裝

```bash
go get github.com/appleboy/easyssh-proxy
```

**需求：** Go 1.24 或更高版本

## 使用方法

你可以在 [`examples`](./_examples/) 資料夾中看到 `ssh`、`scp`、`Proxy` 和 `stream` 命令的詳細範例。

### MakeConfig

這個套件提供的所有功能都是通過 MakeConfig 結構體的方法來訪問的。

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

MakeConfig 接受以下屬性：

| 屬性              | 描述                                                                         |
| ----------------- | ---------------------------------------------------------------------------- |
| user              | 要登入的 SSH 用戶                                                            |
| Server            | 伺服器的 IP 或主機名稱                                                       |
| Key               | 包含用於建立連接的私鑰的字串                                                 |
| KeyPath           | 指向用於建立連接的 SSH 密鑰文件的路徑                                        |
| Port              | 連接到伺服器的 SSH 守護程序時使用的端口                                      |
| Protocol          | 要使用的 TCP 協議："tcp", "tcp4", "tcp6"                                     |
| Passphrase        | 用於解鎖提供的 SSH 密鑰的密碼（如果不需要密碼，則留空）                      |
| Password          | 用於登入指定用戶的密碼                                                       |
| Timeout           | 請求超時前等待的時間長度                                                     |
| Proxy             | 一組額外的配置參數，將通過此頂層塊中配置的伺服器 SSH 到另一個伺服器          |
| Ciphers           | 用於 SSH 連接的密碼陣列（例如 aes256-ctr）                                   |
| KeyExchanges      | 用於 SSH 連接的密鑰交換陣列（例如 ecdh-sha2-nistp384）                       |
| Fingerprint       | SSH 伺服器返回的預期指紋，如果不匹配則會導致指紋錯誤                         |
| UseInsecureCipher | 啟用不安全的密碼和密鑰交換，這些是不安全的，可能會導致妥協，[參見 ssh](#ssh) |

注意：請查看參考文件以獲取 [MakeConfig](https://pkg.go.dev/github.com/appleboy/easyssh-proxy#MakeConfig) 和 [DefaultConfig](https://pkg.go.dev/github.com/appleboy/easyssh-proxy#DefaultConfig) 的最新屬性。

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
  // Please make sure the `tmp` folder exists.
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

注意：代理連接的屬性不會從跳板機繼承。您必須在 DefaultConfig 結構體中明確指定它們。

例如，必須為跳板機（中介伺服器）和目標伺服器分別指定自定義的 `Timeout` 長度。

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

### WriteFile

參見 [examples/writeFile/writeFile.go](./_examples/writeFile/writeFile.go)

```go
func (ssh_conf *MakeConfig) WriteFile(reader io.Reader, size int64, etargetFile string) error
```

```go
func main() {
  // 使用遠程用戶名、伺服器地址和私鑰路徑創建 MakeConfig 實例。
  ssh := &easyssh.MakeConfig{
    Server:  "localhost",
    User:    "drone-scp",
    KeyPath: "./tests/.ssh/id_rsa",
    Port:    "22",
    Timeout: 60 * time.Second,
  }

  fileContents := "Example Text..."
  reader := strings.NewReader(fileContents)

  // 使用 writeFile 命令將文件寫入到遠程伺服器。
  // 第二個參數指定從 reader 中寫入到伺服器的字節數。
  if err := ssh.WriteFile(reader, int64(len(fileContents)), "/home/user/foo.txt"); err != nil {
    return fmt.Errorf("錯誤：無法將文件寫入到客戶端。錯誤：%w", err)
  }
}
```

| 屬性        | 描述                                          |
| ----------- | --------------------------------------------- |
| reader      | 將讀取其內容並保存到伺服器的 `io.reader`      |
| size        | 要從 `io.reader` 中讀取的字節數               |
| etargetFile | 文件將被寫入到伺服器上的位置                   |
