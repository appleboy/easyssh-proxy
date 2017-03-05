# easyssh-proxy

[![GoDoc](https://godoc.org/github.com/appleboy/easyssh-proxy?status.svg)](https://godoc.org/github.com/appleboy/easyssh-proxy) [![Build Status](http://drone.wu-boy.com/api/badges/appleboy/easyssh-proxy/status.svg)](http://drone.wu-boy.com/appleboy/easyssh-proxy) [![codecov](https://codecov.io/gh/appleboy/easyssh-proxy/branch/master/graph/badge.svg)](https://codecov.io/gh/appleboy/easyssh-proxy) [![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/easyssh-proxy)](https://goreportcard.com/report/github.com/appleboy/easyssh-proxy) [![Sourcegraph](https://sourcegraph.com/github.com/appleboy/easyssh-proxy/-/badge.svg)](https://sourcegraph.com/github.com/appleboy/easyssh-proxy?badge)

easyssh-proxy provides a simple implementation of some SSH protocol features in Go.

## Feature

This project is forked from [easyssh](https://github.com/hypersleep/easyssh) but add some features as the following.

* Support plain text of user private key.
* Support key path of user private key.
* Support SSH ProxyCommand.
* Support Timeout for the TCP connection to establish.
