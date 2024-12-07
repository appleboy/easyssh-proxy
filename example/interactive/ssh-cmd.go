package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/appleboy/easyssh-proxy"
	"github.com/ogier/pflag"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	server := pflag.StringP("server", "s", "localhost", "Server to SSH To.")
	scriptFile := pflag.StringP("script", "S", "", "Script to run on remote server")
	username := pflag.StringP("user", "u", os.Getenv("USER"), "Username")
	keyfile := pflag.StringP("key", "k", os.Getenv("HOME")+"/.ssh/id_rsa", "SSH Key")
	help := pflag.BoolP("help", "h", false, "Help")
	pflag.Parse()
	if *help {
		pflag.PrintDefaults()
		os.Exit(1)
	}
	ssh := &easyssh.MakeConfig{
		Server:  *server,
		User:    *username,
		KeyPath: *keyfile,
		Port:    "22",
		Timeout: 60 * time.Second,
	}
	content, err := ioutil.ReadFile(*scriptFile)
	// Call Run method with command you want to run on remote server.
	stdoutChan, stderrChan, doneChan, errChan, err := ssh.Stream(string(content), 60*time.Second)
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
