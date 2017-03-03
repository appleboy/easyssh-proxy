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
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// MakeConfig Contains main authority information.
// User field should be a name of user on remote server (ex. john in ssh john@example.com).
// Server field should be a remote machine address (ex. example.com in ssh john@example.com)
// Key is a path to private key on your local machine.
// Port is SSH server port on remote machine.
// Note: easyssh looking for private key in user's home directory (ex. /home/john + Key).
// Then ensure your Key begins from '/' (ex. /.ssh/id_rsa)
type MakeConfig struct {
	User     string
	Server   string
	Key      string
	KeyPath  string
	Port     string
	Password string
	Timeout  time.Duration
}

// returns ssh.Signer from user you running app home path + cutted key path.
// (ex. pubkey,err := getKeyFile("/.ssh/id_rsa") )
func getKeyFile(keypath string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}

	pubkey, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}

	return pubkey, nil
}

// connects to remote server using MakeConfig struct and returns *ssh.Session
func (ssh_conf *MakeConfig) connect() (*ssh.Session, error) {
	// auths holds the detected ssh auth methods
	auths := []ssh.AuthMethod{}

	// figure out what auths are requested, what is supported
	if ssh_conf.Password != "" {
		auths = append(auths, ssh.Password(ssh_conf.Password))
	}

	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		defer sshAgent.Close()
	}

	if ssh_conf.KeyPath != "" {
		if pubkey, err := getKeyFile(ssh_conf.KeyPath); err == nil {
			auths = append(auths, ssh.PublicKeys(pubkey))
		}
	}

	if ssh_conf.Key != "" {
		signer, _ := ssh.ParsePrivateKey([]byte(ssh_conf.Key))
		auths = append(auths, ssh.PublicKeys(signer))
	}

	config := &ssh.ClientConfig{
		Timeout: ssh_conf.Timeout,
		User:    ssh_conf.User,
		Auth:    auths,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(ssh_conf.Server, ssh_conf.Port), config)
	if err != nil {
		return nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Stream returns one channel that combines the stdout and stderr of the command
// as it is run on the remote machine, and another that sends true when the
// command is done. The sessions and channels will then be closed.
func (ssh_conf *MakeConfig) Stream(command string, timeout int) (stdout chan string, stderr chan string, done chan bool, err error) {
	// connect to remote host
	session, err := ssh_conf.connect()
	if err != nil {
		return stdout, stderr, done, err
	}
	// connect to both outputs (they are of type io.Reader)
	outReader, err := session.StdoutPipe()
	if err != nil {
		return stdout, stderr, done, err
	}
	errReader, err := session.StderrPipe()
	if err != nil {
		return stdout, stderr, done, err
	}
	// combine outputs, create a line-by-line scanner
	stdoutReader := io.MultiReader(outReader)
	stderrReader := io.MultiReader(errReader)
	err = session.Start(command)
	stdoutScanner := bufio.NewScanner(stdoutReader)
	stderrScanner := bufio.NewScanner(stderrReader)
	// continuously send the command's output over the channel
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	done = make(chan bool)

	go func(stdoutScanner, stderrScanner *bufio.Scanner, stdoutChan, stderrChan chan string, done chan bool) {
		defer close(stdoutChan)
		defer close(stderrChan)
		defer close(done)

		timeoutChan := time.After(time.Duration(timeout) * time.Second)
		res := make(chan bool, 1)

		go func() {
			for stdoutScanner.Scan() {
				stdoutChan <- stdoutScanner.Text()
			}
			for stderrScanner.Scan() {
				stderrChan <- stderrScanner.Text()
			}
			// close all of our open resources
			res <- true
		}()

		select {
		case <-res:
			stdoutChan <- ""
			stderrChan <- ""
			done <- true
		case <-timeoutChan:
			stdoutChan <- ""
			stderrChan <- "Run Command Timeout!"
			done <- false
		}

		session.Close()
	}(stdoutScanner, stderrScanner, stdoutChan, stderrChan, done)
	return stdoutChan, stderrChan, done, err
}

// Run command on remote machine and returns its stdout as a string
func (ssh_conf *MakeConfig) Run(command string, timeout int) (outStr string, errStr string, isTimeout bool, err error) {
	stdoutChan, stderrChan, doneChan, err := ssh_conf.Stream(command, timeout)
	if err != nil {
		return outStr, errStr, isTimeout, err
	}
	// read from the output channel until the done signal is passed
	stillGoing := true
	for stillGoing {
		select {
		case isTimeout = <-doneChan:
			stillGoing = false
		case outline := <-stdoutChan:
			if outline != "" {
				outStr += outline + "\n"
			}
		case errline := <-stderrChan:
			if errline != "" {
				errStr += errline + "\n"
			}
		}
	}
	// return the concatenation of all signals from the output channel
	return outStr, errStr, isTimeout, err
}

// Scp uploads sourceFile to remote machine like native scp console app.
func (ssh_conf *MakeConfig) Scp(sourceFile string, etargetFile string) error {
	session, err := ssh_conf.connect()

	if err != nil {
		return err
	}
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

	go func() {
		w, _ := session.StdinPipe()

		fmt.Fprintln(w, "C0644", srcStat.Size(), targetFile)

		if srcStat.Size() > 0 {
			io.Copy(w, src)
			fmt.Fprint(w, "\x00")
			w.Close()
		} else {
			fmt.Fprint(w, "\x00")
			w.Close()
		}
	}()

	if err := session.Run(fmt.Sprintf("scp -tr %s", etargetFile)); err != nil {
		return err
	}

	return nil
}
