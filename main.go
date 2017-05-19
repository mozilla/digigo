// Package digicert is a Go client to the Digicert REST API
// https://www.digicert.com/services/v2/documentation/

package digigo // import "go.mozilla.org/digigo"

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	r.Header.Set("User-Agent", "Digicert Go Client "+Version)
	r.Header.Set("X-DC-DEVKEY", cli.token)
	r.Header.Set("Accept", "application/json")
	if cli.debug {
		fmt.Printf("debug: %s %s %s\ndebug: User-Agent: %s\ndebug: X-DC-DEVKEY: %s\n",
			r.Method, r.URL.String(), r.Proto, r.UserAgent(), r.Header.Get("X-DC-DEVKEY"))
	}
	// execute the request
	resp, err := cli.requester.Do(r)
	if resp == nil {
		return nil, errors.New("received empty response from digicert api")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to request the digicert api: %d %s", resp.StatusCode, resp.Status)
	}
	if resp.StatusCode > 200 {
		return nil, errors.Errorf("failed to request the digicert api: %d %s", resp.StatusCode, resp.Status)
	}
	return resp, nil
}

// Product defines a product that can be ordered
type Product struct {
	GroupName string `json:"group_name"`
	NameID    string `json:"name_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

type productList struct {
	Products []Product `json:"products"`
}

// ViewProductList returns a list of Products
func (cli Client) ViewProductList() ([]Product, error) {
	r, err := http.NewRequest("GET", cli.baseurl+"/product", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve list of products from digicert api")
	}
	resp, err := cli.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve list of products from digicert api")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	defer resp.Body.Close()
	var pl productList
	err = json.Unmarshal(body, &pl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse JSON response body")
	}
	return pl.Products, nil
}