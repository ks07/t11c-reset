/*
Copyright © 2020 George Field <george@cucurbit.dev>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package t11c

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"golang.org/x/net/publicsuffix"
)

type Connection struct {
	DryRun   bool // If true, don't make any changes to the modem
	Username string
	Password string
	Hostname string
	client   *http.Client
	logger   log.Logger
}

func NewConnection(logger log.Logger, dryrun bool, username, password, hostname string) *Connection {
	return &Connection{
		DryRun:   dryrun,
		Username: username,
		Password: password,
		Hostname: hostname,
		logger:   logger,
	}
}

func (c *Connection) init() error {
	level.Debug(c.logger).Log("msg", "initialising client")
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return err
	}

	c.client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// We don't want to follow any redirects automatically
			return http.ErrUseLastResponse
		},
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	return nil
}

func (c *Connection) getURL(path string) url.URL {
	return url.URL{
		Scheme: "http",
		Host:   c.Hostname,
		Path:   path,
	}
}

func (c *Connection) ignoreBody(resp *http.Response) error {
	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return err
	}
	return resp.Body.Close()
}

func (c *Connection) getWithContext(ctx context.Context, u url.URL) (*http.Response, error) {
	level.Debug(c.logger).Log("request_url", u.String(), "msg", "requesting URL")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req)
}

func (c *Connection) postFormWithContext(ctx context.Context, u url.URL, data url.Values) (*http.Response, error) {
	level.Debug(c.logger).Log("request_url", u.String(), "msg", "submitting form")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.client.Do(req)
}

func (c *Connection) Login(ctx context.Context) error {
	if c.client == nil {
		if err := c.init(); err != nil {
			return err
		}
	}

	// A session cookie is only assigned on the 302 to the login page
	initURL := c.getURL("/")
	initResp, err := c.getWithContext(ctx, initURL)
	if err != nil {
		return err
	}
	if err := c.ignoreBody(initResp); err != nil {
		return err
	}

	// The T11C feeds the credentials as a base64 query string parameter, in a GET request...
	credParam := fmt.Sprintf("%s:%s", c.Username, c.Password)
	credParamEncoded := base64.StdEncoding.EncodeToString([]byte(credParam))

	loginURL := c.getURL("/cgi-bin/index.asp")
	// The T11C doesn't pass this as a value, and doesn't escape any trailing '='!
	loginURL.RawQuery = credParamEncoded

	loginResp, err := c.getWithContext(ctx, loginURL)
	if err != nil {
		return err
	}
	return c.ignoreBody(loginResp)
}

func (c *Connection) TestSession(ctx context.Context) (bool, error) {
	if c.client == nil {
		if err := c.init(); err != nil {
			return false, err
		}
	}

	u := c.getURL("/cgi-bin/main.html")
	resp, err := c.getWithContext(ctx, u)
	if err != nil {
		return false, err
	}
	if err = c.ignoreBody(resp); err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (c *Connection) ModemIsConnected(ctx context.Context) (bool, error) {
	if c.client == nil {
		if err := c.init(); err != nil {
			return false, err
		}
	}

	u := c.getURL("/cgi-bin/pages/statusview.cgi")
	resp, err := c.getWithContext(ctx, u)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	ip, err := extractWANIP(resp.Body)
	if errors.Is(err, errWANIPTextNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return ip != "0.0.0.0", nil
}

func (c *Connection) SetModemState(ctx context.Context, connect bool) error {
	if c.client == nil {
		err := c.init()
		if err != nil {
			return err
		}
	}

	if c.DryRun {
		return nil
	}

	// The typo here is intentional
	u := c.getURL("/cgi-bin/PPPoEManulDial.asp")

	data := url.Values{}
	data.Add("Dipflag", "0")
	// The redirect flag is passed as 0 for disconnects and 1 for connects by the web interface,
	// although it doesn't appear to do anything...
	data.Add("redirect", "0")
	if connect {
		data.Add("DipConnFlag", "1")
	} else {
		data.Add("DipConnFlag", "2")
	}

	resp, err := c.postFormWithContext(ctx, u, data)
	if err != nil {
		return err
	}
	return c.ignoreBody(resp)
}
