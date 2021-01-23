package easyssh

import (
	"bufio"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

type Direct struct{}

func (Direct) Dial(network, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
}

type connectProxyDialer struct {
	host     string
	forward  proxy.Dialer
	auth     bool
	username string
	password string
}

func NewConnectProxyDialer(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
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

func RegisterDialerType() {
	proxy.RegisterDialerType("http", NewConnectProxyDialer)
	proxy.RegisterDialerType("https", NewConnectProxyDialer)
}

func NewHttpProxyConn(d Direct, proxyAddr, targetAddr string) (net.Conn, error) {
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

func (p *connectProxyDialer) Dial(network, addr string) (net.Conn, error) {
	c, err := p.forward.Dial("tcp", p.host)

	if err != nil {
		return nil, err
	}

	reqUrl, err := url.Parse("http://" + addr)
	if err != nil {
		c.Close()
		return nil, err
	}

	req, err := http.NewRequest("CONNECT", reqUrl.String(), nil)
	if err != nil {
		c.Close()
		return nil, err
	}

	if p.auth {
		req.SetBasicAuth(p.username, p.password)
	}

	req.Close = false

	err = req.Write(c)
	if err != nil {
		c.Close()
		return nil, err
	}

	res, err := http.ReadResponse(bufio.NewReader(c), req)

	if err != nil {
		res.Body.Close()
		c.Close()
		return nil, err
	}

	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		c.Close()
		return nil, fmt.Errorf("Connection Error: StatusCode: %d", res.StatusCode)
	}

	return c, nil
}
