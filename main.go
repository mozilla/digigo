// Package digigo is a Go client to the Digicert REST API
// https://www.digicert.com/services/v2/documentation/
package digigo // import "go.mozilla.org/digigo"

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

// Version is the version of this library
const Version string = "0.0.1"

// Client is a generic API client
type Client struct {
	requester *http.Client
	token     string
	debug     bool
	baseurl   string
}

// NewClient initiates a new instance of a Client
func NewClient(token string) (cli Client, err error) {
	cli.token = token
	tr := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  false,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			InsecureSkipVerify: false,
		},
		Proxy: http.ProxyFromEnvironment,
	}
	cli.requester = &http.Client{Transport: tr}
	cli.baseurl = "https://www.digicert.com/services/v2"
	return
}

// ChangeBaseURL resets the base URL of the Digicert REST API to a provided value
func (cli *Client) ChangeBaseURL(baseurl string) {
	cli.baseurl = baseurl
	return
}

// EnableDebug will print extra information when contacting Digicert
func (cli *Client) EnableDebug() {
	cli.debug = true
	return
}

// DisableDebug will hide extra information when contacting Digicert
func (cli *Client) DisableDebug() {
	cli.debug = false
	return
}

// Do is a thin wrapper around http.Client.Do() that inserts an authentication header
// to the outgoing request and checks response codes
func (cli Client) Do(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", "go.mozilla.org/digigo "+Version)
	r.Header.Set("X-DC-DEVKEY", cli.token)
	r.Header.Set("Accept", "application/json")
	if r.Method == http.MethodPost && r.Body != nil {
		// POST Body is always JSON, so set the content type
		r.Header.Set("Content-Type", "application/json")
	}
	if cli.debug {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Fatal("failed to dump request in debug mode: ", err)
		}
		color.Green("--- debug: client request ---\n%s\n-----------------------------\n", dump)
	}
	// execute the request
	resp, err := cli.requester.Do(r)
	if resp == nil {
		return nil, errors.New("received empty response from digicert api")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to request the digicert api: %s", resp.Status)
	}
	if cli.debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal("failed to dump response in debug mode: ", err)
		}
		color.Red("--- debug: digicert response ---\n%s\n--------------------------------\n", dump)
	}

	if resp.StatusCode >= 300 {
		if resp.Body != nil {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, errors.Errorf("failed to request the digicert api: %s. couldn't parse returned error %q: %s",
					resp.Status, body, err)
			}
			var errs Errors
			err = json.Unmarshal(body, &errs)
			if err != nil {
				return nil, errors.Errorf("failed to request the digicert api: %s. couldn't parse returned error %q: %s",
					resp.Status, body, errs)
			}
			return nil, errors.Errorf("failed to request the digicert api: %s, %s", resp.Status, errs)
		}
		return nil, errors.Errorf("failed to request the digicert api: %s. no error was returned.", resp.Status)
	}
	return resp, nil
}
