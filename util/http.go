package util

import (
	"bytes"
	"context"
	"errors"
	//"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"git.drinkme.beer/yinghe/log"
)

// Network contains common configuration.
type Network struct {
	Env            string   `toml:"env"`
	Host           []string `toml:"gateway"`
	GatewayURL     string
	RequestTimeout time.Duration `toml:"request_timeout"`
	ConnectTimeout time.Duration `toml:"connect_timeout"`
	SocketTimeout  time.Duration `toml:"socket_timeout"`
}

var network = Network{
	RequestTimeout: 45 * time.Second,
	ConnectTimeout: 45 * time.Second,
	SocketTimeout:  55 * time.Second,
}

var client = &http.Client{
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, n, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(n, addr,
				time.Second*network.RequestTimeout,
			)
			if err != nil {
				return conn, err
			}
			conn.SetDeadline(time.Now().
				Add(time.Second * network.ConnectTimeout))
			return conn, err
		},
		ResponseHeaderTimeout: time.Second * network.SocketTimeout,
		MaxConnsPerHost:       20,
		IdleConnTimeout:       3 * time.Minute,
		MaxIdleConns:          10,
		//Proxy: func(_ *http.Request) (*url.URL, error) {
		//	return url.Parse("http://192.168.0.104:1087")
		//},
	},
}

func SetNetworkCfg(cfg Network) {
	if 0 < cfg.RequestTimeout {
		network.RequestTimeout = cfg.RequestTimeout
	}

	if 0 < cfg.ConnectTimeout {
		network.ConnectTimeout = cfg.ConnectTimeout
	}

	if 0 < cfg.SocketTimeout {
		network.SocketTimeout = cfg.SocketTimeout
	}
}

type Body interface {
	BuildRequest() []byte
}

type Client struct {
	GatewayURL    string
	URI           string
	Headers       map[string]string
	Authorization string
	Method        string
	HTTPStatus    int
	Body          []byte // Indicates both Request Body & Response Body
	TraceId       interface{}
	//Buffer        *bufio.Reader
}

func (c *Client) AddHeader(k, v string) {
	c.Headers[k] = v
}

func (c *Client) Print() {
	if nil != c.Body {
		log.Infof(string(c.Body))
	} else {
		log.Infof("nil")
	}
}

type Bytes []byte

func (self Bytes) BuildRequest() []byte {
	return self
}

/*
url, uri, method, body
*/
func NewHttpClient(url, uri, m string, b []byte) *Client {
	return &Client{
		GatewayURL: url,
		URI:        uri,
		Method:     m,
		Headers:    make(map[string]string),
		Body:       b,
	}
}

func (c *Client) BuildRequest(body Body) {
	if nil == c.Body {
		c.Body = make([]byte, 0, 1024)
	}
	c.Body = append(c.Body, body.BuildRequest()...)
}

func (c *Client) Request() (err error) {
	log.Infof("request url: %s", c.GatewayURL+c.URI)
	request, err := http.NewRequest(c.Method,
		c.GatewayURL+c.URI,
		bytes.NewReader(c.Body),
	)
	if err != nil {
		log.Info("Post err occurs:", err.Error())
		return
	}

	for k, v := range c.Headers {
		request.Header.Set(k, v)
	}

	resp, err := client.Do(request)
	if nil == resp {
		log.Info("none response received")
		err = errors.New("none response received")
		return
	}

	defer resp.Body.Close()
	c.HTTPStatus = resp.StatusCode
	c.Body, err = ioutil.ReadAll(resp.Body)

	return
}

func Request(req *http.Request) ([]byte, error) {
	var (
		err  error
		body []byte
	)

	if nil == req {
		log.Errorf("illegal request")
		return nil, errors.New("illegal request")
	}
	//log.Infof("%v", req)
	resp, err := client.Do(req)
	if nil == resp {
		log.Info("none response received")
		err = errors.New("none response received")
		return nil, err
	}

	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		if nil != resp.Body {

			body, _ = ioutil.ReadAll(resp.Body)
			log.Infof("response: %s", string(body))
		}
		log.Errorf("unexpected status: %d", resp.StatusCode)
		return nil, errors.New("unexpected status")
	}

	body, err = ioutil.ReadAll(resp.Body)
	if nil != err {
		log.Errorf("invalid response body: %s", err.Error())
		return nil, err
	}

	log.Infof("response: %s", string(body))

	return body, nil
}
