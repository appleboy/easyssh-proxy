# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

easyssh-proxy is a Go library that provides a simple SSH client implementation with support for SSH tunneling/proxy connections. It's forked from the original easyssh project with additional features for proxy connections, timeout handling, and secure key management.

## Core Architecture

### Main Types

- **MakeConfig**: Primary configuration struct containing SSH connection parameters (user, server, keys, timeouts, proxy settings)
- **DefaultConfig**: Configuration struct used for SSH proxy/jumphost connections
- Both structs share similar fields but MakeConfig includes additional proxy capabilities

### Key Methods

- `Connect()`: Establishes SSH session and client connection
- `Run()`: Executes single command and returns output
- `Stream()`: Executes command with real-time streaming output via channels
- `Scp()`: Copies files to remote server
- `WriteFile()`: Writes content from io.Reader to remote file

### Authentication Support

- Password authentication
- Private key files (with optional passphrase)
- Raw private key content (embedded in code)
- SSH agent integration
- Custom cipher and key exchange algorithms

### Proxy/Jumphost Architecture

The library supports SSH proxy connections where traffic is tunneled through an intermediate server:

```text
Client -> Jumphost -> Target Server
```

The `Proxy` field in MakeConfig uses DefaultConfig to define the jumphost connection parameters.

## Development Commands

### Testing

```bash
make test                    # Run all tests with coverage
go test -v ./...            # Run tests verbose
```

### Code Quality

```bash
make fmt                     # Format code using gofumpt
make vet                     # Run go vet
```

### SSH Test Server Setup

```bash
make ssh-server             # Setup local SSH test server (Alpine Linux)
```

### Linting

The project uses golangci-lint via GitHub Actions. No local lint command is defined in the Makefile.

## Testing Infrastructure

### Test Environment

- Uses Alpine Linux container with SSH server setup
- Creates test users: `drone-scp` and `root`
- SSH keys located in `tests/.ssh/` directory
- Test files in `tests/` include sample data and configuration

### CI/CD

- GitHub Actions workflow in `.github/workflows/testing.yml`
- Runs tests in Go 1.23 Alpine container
- Includes golangci-lint for code quality
- Codecov integration for coverage reporting

## Code Patterns

### Configuration Pattern

```go
ssh := &easyssh.MakeConfig{
    User:    "username",
    Server:  "hostname",
    KeyPath: "/path/to/key",
    Port:    "22",
    Timeout: 60 * time.Second,
}
```

### Proxy Configuration

```go
ssh := &easyssh.MakeConfig{
    // ... main server config
    Proxy: easyssh.DefaultConfig{
        // ... jumphost config
    },
}
```

### Error Handling

Functions return multiple values following Go conventions:

- `Run()`: (stdout, stderr, isTimeout, error)
- `Stream()`: (stdoutChan, stderrChan, doneChan, errChan, error)

## Example Usage

Comprehensive examples are available in `_examples/` directory:

- `ssh/`: Basic command execution
- `scp/`: File copying
- `proxy/`: SSH tunneling through jumphost
- `stream/`: Real-time command output streaming
- `writeFile/`: Writing content to remote files

## Dependencies

- `golang.org/x/crypto/ssh`: Core SSH protocol implementation
- `github.com/ScaleFT/sshkeys`: Enhanced private key parsing with passphrase support
- `github.com/stretchr/testify`: Testing framework

## Security Considerations

- Supports both secure and insecure cipher configurations
- `UseInsecureCipher` flag enables legacy/weak ciphers when needed
- Fingerprint verification available for enhanced security
- Private keys can be embedded as strings or loaded from files
