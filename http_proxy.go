package easyssh

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

type directDialer struct{}

func (directDialer) Dial(network, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
}

type connectProxyDialer struct {
	host     string
	forward  proxy.Dialer
	auth     bool
	username string
	password string
}

func newConnectProxyDialer(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	host := u.Host
	p := &connectProxyDialer{
		host:    host,
		forward: forward,
	}

	if u.User != nil {
		p.auth = true
		p.username = u.User.Username()
		p.password, _ = u.User.Password()
	}

	return p, nil
}

func registerDialerType() {
	proxy.RegisterDialerType("http", newConnectProxyDialer)
	proxy.RegisterDialerType("https", newConnectProxyDialer)
}

func newHTTPProxyConn(d directDialer, proxyAddr, targetAddr string) (net.Conn, error) {
	proxyURL, err := url.Parse("http://" + proxyAddr)
	if err != nil {
		return nil, err
	}

	proxyDialer, err := proxy.FromURL(proxyURL, d)
	if err != nil {
		return nil, err
	}

	proxyConn, err := proxyDialer.Dial("tcp", targetAddr)
	if err != nil {
		return nil, err
	}

	return proxyConn, err
}

func (p *connectProxyDialer) Dial(_, addr string) (net.Conn, error) {
	c, err := p.forward.Dial("tcp", p.host)
	if err != nil {
		return nil, err
	}

	reqURL, err := url.Parse("http://" + addr)
	if err != nil {
		_ = c.Close()
		return nil, err
	}

	req, err := http.NewRequest("CONNECT", reqURL.String(), nil)
	if err != nil {
		_ = c.Close()
		return nil, err
	}

	if p.auth {
		req.SetBasicAuth(p.username, p.password)
	}

	req.Close = false

	err = req.Write(c)
	if err != nil {
		_ = c.Close()
		return nil, err
	}

	res, err := http.ReadResponse(bufio.NewReader(c), req)
	if err != nil {
		res.Body.Close()
		_ = c.Close()
		return nil, err
	}

	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		_ = c.Close()
		return nil, fmt.Errorf("Connection Error: StatusCode: %d", res.StatusCode)
	}

	return c, nil
}
