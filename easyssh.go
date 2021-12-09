// Package easyssh provides a simple implementation of some SSH protocol
// features in Go. You can simply run a command on a remote server or get a file
// even simpler than native console SSH client. You don't need to think about
// Dials, sessions, defers, or public keys... Let easyssh think about it!
package easyssh

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ScaleFT/sshkeys"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var defaultTimeout = 60 * time.Second

type (
	// MakeConfig Contains main authority information.
	// User field should be a name of user on remote server (ex. john in ssh john@example.com).
	// Server field should be a remote machine address (ex. example.com in ssh john@example.com)
	// Key is a path to private key on your local machine.
	// Port is SSH server port on remote machine.
	// Note: easyssh looking for private key in user's home directory (ex. /home/john + Key).
	// Then ensure your Key begins from '/' (ex. /.ssh/id_rsa)
	MakeConfig struct {
		User         string
		Server       string
		Key          string
		KeyPath      string
		Port         string
		Passphrase   string
		Password     string
		Timeout      time.Duration
		Proxy        DefaultConfig
		Ciphers      []string
		KeyExchanges []string
		Fingerprint  string

		// Enable the use of insecure ciphers and key exchange methods.
		// This enables the use of the the following insecure ciphers and key exchange methods:
		// - aes128-cbc
		// - aes192-cbc
		// - aes256-cbc
		// - 3des-cbc
		// - diffie-hellman-group-exchange-sha256
		// - diffie-hellman-group-exchange-sha1
		// Those algorithms are insecure and may allow plaintext data to be recovered by an attacker.
		UseInsecureCipher bool
	}

	// DefaultConfig for ssh proxy config
	DefaultConfig struct {
		User         string
		Server       string
		Key          string
		KeyPath      string
		Port         string
		Passphrase   string
		Password     string
		Timeout      time.Duration
		Ciphers      []string
		KeyExchanges []string
		Fingerprint  string

		// Enable the use of insecure ciphers and key exchange methods.
		// This enables the use of the the following insecure ciphers and key exchange methods:
		// - aes128-cbc
		// - aes192-cbc
		// - aes256-cbc
		// - 3des-cbc
		// - diffie-hellman-group-exchange-sha256
		// - diffie-hellman-group-exchange-sha1
		// Those algorithms are insecure and may allow plaintext data to be recovered by an attacker.
		UseInsecureCipher bool
	}
)

// returns ssh.Signer from user you running app home path + cutted key path.
// (ex. pubkey,err := getKeyFile("/.ssh/id_rsa") )
func getKeyFile(keypath, passphrase string) (ssh.Signer, error) {
	var pubkey ssh.Signer
	var err error
	buf, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}

	if passphrase != "" {
		pubkey, err = sshkeys.ParseEncryptedPrivateKey(buf, []byte(passphrase))
	} else {
		pubkey, err = ssh.ParsePrivateKey(buf)
	}

	if err != nil {
		return nil, err
	}

	return pubkey, nil
}

// returns *ssh.ClientConfig and io.Closer.
// if io.Closer is not nil, io.Closer.Close() should be called when
// *ssh.ClientConfig is no longer used.
func getSSHConfig(config DefaultConfig) (*ssh.ClientConfig, io.Closer) {
	var sshAgent io.Closer

	// auths holds the detected ssh auth methods
	auths := []ssh.AuthMethod{}

	// figure out what auths are requested, what is supported
	if config.Password != "" {
		auths = append(auths, ssh.Password(config.Password))
	}
	if config.KeyPath != "" {
		if pubkey, err := getKeyFile(config.KeyPath, config.Passphrase); err != nil {
			log.Printf("getKeyFile error: %v\n", err)
		} else {
			auths = append(auths, ssh.PublicKeys(pubkey))
		}
	}

	if config.Key != "" {
		var signer ssh.Signer
		var err error
		if config.Passphrase != "" {
			signer, err = sshkeys.ParseEncryptedPrivateKey([]byte(config.Key), []byte(config.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(config.Key))
		}

		if err != nil {
			log.Printf("ssh.ParsePrivateKey: %v\n", err)
		} else {
			auths = append(auths, ssh.PublicKeys(signer))
		}
	}

	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}

	c := ssh.Config{}
	if config.UseInsecureCipher {
		c.SetDefaults()
		c.Ciphers = append(c.Ciphers, "aes128-cbc", "aes192-cbc", "aes256-cbc", "3des-cbc")
		c.KeyExchanges = append(c.KeyExchanges, "diffie-hellman-group-exchange-sha1", "diffie-hellman-group-exchange-sha256")
	}

	if len(config.Ciphers) > 0 {
		c.Ciphers = append(c.Ciphers, config.Ciphers...)
	}

	if len(config.KeyExchanges) > 0 {
		c.KeyExchanges = append(c.KeyExchanges, config.KeyExchanges...)
	}

	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	if config.Fingerprint != "" {
		hostKeyCallback = func(hostname string, remote net.Addr, publicKey ssh.PublicKey) error {
			if ssh.FingerprintSHA256(publicKey) != config.Fingerprint {
				return fmt.Errorf("ssh: host key fingerprint mismatch")
			}
			return nil
		}
	}

	return &ssh.ClientConfig{
		Config:          c,
		Timeout:         config.Timeout,
		User:            config.User,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
	}, sshAgent
}

// Connect to remote server using MakeConfig struct and returns *ssh.Session
func (ssh_conf *MakeConfig) Connect() (*ssh.Session, *ssh.Client, error) {
	var client *ssh.Client
	var err error

	targetConfig, closer := getSSHConfig(DefaultConfig{
		User:              ssh_conf.User,
		Key:               ssh_conf.Key,
		KeyPath:           ssh_conf.KeyPath,
		Passphrase:        ssh_conf.Passphrase,
		Password:          ssh_conf.Password,
		Timeout:           ssh_conf.Timeout,
		Ciphers:           ssh_conf.Ciphers,
		KeyExchanges:      ssh_conf.KeyExchanges,
		Fingerprint:       ssh_conf.Fingerprint,
		UseInsecureCipher: ssh_conf.UseInsecureCipher,
	})
	if closer != nil {
		defer closer.Close()
	}

	// Enable proxy command
	if ssh_conf.Proxy.Server != "" {
		proxyConfig, closer := getSSHConfig(DefaultConfig{
			User:              ssh_conf.Proxy.User,
			Key:               ssh_conf.Proxy.Key,
			KeyPath:           ssh_conf.Proxy.KeyPath,
			Passphrase:        ssh_conf.Proxy.Passphrase,
			Password:          ssh_conf.Proxy.Password,
			Timeout:           ssh_conf.Proxy.Timeout,
			Ciphers:           ssh_conf.Proxy.Ciphers,
			KeyExchanges:      ssh_conf.Proxy.KeyExchanges,
			Fingerprint:       ssh_conf.Proxy.Fingerprint,
			UseInsecureCipher: ssh_conf.Proxy.UseInsecureCipher,
		})
		if closer != nil {
			defer closer.Close()
		}

		proxyClient, err := ssh.Dial("tcp", net.JoinHostPort(ssh_conf.Proxy.Server, ssh_conf.Proxy.Port), proxyConfig)
		if err != nil {
			return nil, nil, err
		}

		conn, err := proxyClient.Dial("tcp", net.JoinHostPort(ssh_conf.Server, ssh_conf.Port))
		if err != nil {
			return nil, nil, err
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(ssh_conf.Server, ssh_conf.Port), targetConfig)
		if err != nil {
			return nil, nil, err
		}

		client = ssh.NewClient(ncc, chans, reqs)
	} else {
		client, err = ssh.Dial("tcp", net.JoinHostPort(ssh_conf.Server, ssh_conf.Port), targetConfig)
		if err != nil {
			return nil, nil, err
		}
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, nil, err
	}

	return session, client, nil
}

// Stream returns one channel that combines the stdout and stderr of the command
// as it is run on the remote machine, and another that sends true when the
// command is done. The sessions and channels will then be closed.
func (ssh_conf *MakeConfig) Stream(command string, timeout ...time.Duration) (<-chan string, <-chan string, <-chan bool, <-chan error, error) {
	// continuously send the command's output over the channel
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	doneChan := make(chan bool)
	errChan := make(chan error)

	// connect to remote host
	session, client, err := ssh_conf.Connect()
	if err != nil {
		return stdoutChan, stderrChan, doneChan, errChan, err
	}
	// defer session.Close()
	// connect to both outputs (they are of type io.Reader)
	outReader, err := session.StdoutPipe()
	if err != nil {
		client.Close()
		session.Close()
		return stdoutChan, stderrChan, doneChan, errChan, err
	}
	errReader, err := session.StderrPipe()
	if err != nil {
		client.Close()
		session.Close()
		return stdoutChan, stderrChan, doneChan, errChan, err
	}
	err = session.Start(command)
	if err != nil {
		client.Close()
		session.Close()
		return stdoutChan, stderrChan, doneChan, errChan, err
	}

	// combine outputs, create a line-by-line scanner
	stdoutReader := io.MultiReader(outReader)
	stderrReader := io.MultiReader(errReader)
	stdoutScanner := bufio.NewScanner(stdoutReader)
	stderrScanner := bufio.NewScanner(stderrReader)

	go func(stdoutScanner, stderrScanner *bufio.Scanner, stdoutChan, stderrChan chan string, doneChan chan bool, errChan chan error) {
		defer close(doneChan)
		defer close(errChan)
		defer client.Close()
		defer session.Close()

		// default timeout value
		executeTimeout := defaultTimeout
		if len(timeout) > 0 {
			executeTimeout = timeout[0]
		}
		timeoutChan := time.After(executeTimeout)
		res := make(chan struct{}, 1)
		var resWg sync.WaitGroup
		resWg.Add(2)

		go func() {
			defer close(stdoutChan)
			for stdoutScanner.Scan() {
				stdoutChan <- stdoutScanner.Text()
			}
			resWg.Done()
		}()

		go func() {
			defer close(stderrChan)
			for stderrScanner.Scan() {
				stderrChan <- stderrScanner.Text()
			}
			resWg.Done()
		}()

		go func() {
			resWg.Wait()
			// close all of our open resources
			res <- struct{}{}
		}()

		select {
		case <-res:
			errChan <- session.Wait()
			doneChan <- true
		case <-timeoutChan:
			errChan <- fmt.Errorf("Run Command Timeout")
			doneChan <- false
		}
	}(stdoutScanner, stderrScanner, stdoutChan, stderrChan, doneChan, errChan)

	return stdoutChan, stderrChan, doneChan, errChan, err
}

// Run command on remote machine and returns its stdout as a string
func (ssh_conf *MakeConfig) Run(command string, timeout ...time.Duration) (outStr string, errStr string, isTimeout bool, err error) {
	stdoutChan, stderrChan, doneChan, errChan, err := ssh_conf.Stream(command, timeout...)
	if err != nil {
		return outStr, errStr, isTimeout, err
	}
	// read from the output channel until the done signal is passed
loop:
	for {
		select {
		case isTimeout = <-doneChan:
			break loop
		case outline, ok := <-stdoutChan:
			if !ok {
				stdoutChan = nil
			}
			if outline != "" {
				outStr += outline + "\n"
			}
		case errline, ok := <-stderrChan:
			if !ok {
				stderrChan = nil
			}
			if errline != "" {
				errStr += errline + "\n"
			}
		case err = <-errChan:
		}
	}
	// return the concatenation of all signals from the output channel
	return outStr, errStr, isTimeout, err
}

// Scp uploads sourceFile to remote machine like native scp console app.
func (ssh_conf *MakeConfig) Scp(sourceFile string, etargetFile string) error {
	session, client, err := ssh_conf.Connect()
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	targetFile := filepath.Base(etargetFile)

	src, srcErr := os.Open(sourceFile)

	if srcErr != nil {
		return srcErr
	}

	srcStat, statErr := src.Stat()

	if statErr != nil {
		return statErr
	}

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	copyF := func() error {
		_, err := fmt.Fprintln(w, "C0644", srcStat.Size(), targetFile)
		if err != nil {
			return err
		}

		if srcStat.Size() > 0 {
			_, err = io.Copy(w, src)
			if err != nil {
				return err
			}
		}

		_, err = fmt.Fprint(w, "\x00")
		if err != nil {
			return err
		}

		return nil
	}

	copyErrC := make(chan error, 1)
	go func() {
		defer w.Close()
		copyErrC <- copyF()
	}()

	err = session.Run(fmt.Sprintf("scp -tr %s", etargetFile))
	if err != nil {
		return err
	}

	err = <-copyErrC
	return err
}
