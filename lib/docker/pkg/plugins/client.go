package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hyperhq/runv/lib/glog"
)

const (
	versionMimetype = "application/vnd.docker.plugins.v1+json"
	defaultTimeOut  = 30
)

func NewClient(addr string) *Client {
	tr := &http.Transport{}
	protoAndAddr := strings.Split(addr, "://")
	configureTCPTransport(tr, protoAndAddr[0], protoAndAddr[1])
	return &Client{&http.Client{Transport: tr}, protoAndAddr[1]}
}

type Client struct {
	http *http.Client
	addr string
}

func (c *Client) Call(serviceMethod string, args interface{}, ret interface{}) error {
	return c.callWithRetry(serviceMethod, args, ret, true)
}

func (c *Client) callWithRetry(serviceMethod string, args interface{}, ret interface{}, retry bool) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(args); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "/"+serviceMethod, &buf)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", versionMimetype)
	req.URL.Scheme = "http"
	req.URL.Host = c.addr

	var retries int
	start := time.Now()

	for {
		resp, err := c.http.Do(req)
		if err != nil {
			if !retry {
				return err
			}

			timeOff := backoff(retries)
			if abort(start, timeOff) {
				return err
			}
			retries++
			glog.Warningf("Unable to connect to plugin: %s, retrying in %v", c.addr, timeOff)
			time.Sleep(timeOff)
			continue
		}

		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			remoteErr, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("Plugin Error: %s", err)
			}
			return fmt.Errorf("Plugin Error: %s", remoteErr)
		}

		return json.NewDecoder(resp.Body).Decode(&ret)
	}
}

func backoff(retries int) time.Duration {
	b, max := 1, defaultTimeOut
	for b < max && retries > 0 {
		b *= 2
		retries--
	}
	if b > max {
		b = max
	}
	return time.Duration(b) * time.Second
}

func abort(start time.Time, timeOff time.Duration) bool {
	return timeOff+time.Since(start) > time.Duration(defaultTimeOut)*time.Second
}

func configureTCPTransport(tr *http.Transport, proto, addr string) {
	// Why 32? See https://github.com/docker/docker/pull/8035.
	timeout := 32 * time.Second
	if proto == "unix" {
		// No need for compression in local communications.
		tr.DisableCompression = true
		tr.Dial = func(_, _ string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	} else {
		tr.Proxy = http.ProxyFromEnvironment
		tr.Dial = (&net.Dialer{Timeout: timeout}).Dial
	}
}