package digigo // import "go.mozilla.org/digigo"

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

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
		return nil, errors.Wrap(err, "failed to prepare request to list products")
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
