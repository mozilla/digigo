package digigo // import "go.mozilla.org/digigo"
import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// Order is the structure used to submit a certificate order to Digicert.
// https://www.digicert.com/services/v2/documentation/order/overview-submit
type Order struct {
	Certificate struct {
		CommonName        string   `json:"common_name"`
		DNSNames          []string `json:"dns_names"`
		Csr               string   `json:"csr"`
		OrganizationUnits []string `json:"organization_units,omitempty"`
		ServerPlatform    struct {
			ID int `json:"id"`
		} `json:"server_platform,omitempty"`
		SignatureHash string `json:"signature_hash"`
		ProfileOption string `json:"profile_option,omitempty"`
	} `json:"certificate"`
	Organization struct {
		ID int `json:"id"`
	} `json:"organization"`
	ValidityYears               int    `json:"validity_years"`
	CustomExpirationDate        string `json:"custom_expiration_date,omitempty"`
	Comments                    string `json:"comments,omitempty"`
	DisableRenewalNotifications bool   `json:"disable_renewal_notifications,omitempty"`
	RenewalOfOrderID            int    `json:"renewal_of_order_id,omitempty"`
	PaymentMethod               string `json:"payment_method,omitempty"`
}

type orderResponse struct {
	ID       int `json:"id"`
	Requests []struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	} `json:"requests"`
}

// SubmitOrder sends an order for a given product to the Digicert API. The productNameId
// determines which product is being ordered. The full list of productNameId an account has
// access to can be obtained via ViewProductList().
//
// This function returns the orderId, or -1 and an error if not successful.
func (cli Client) SubmitOrder(order Order, productNameID string) (int, error) {
	orderBody, err := json.Marshal(order)
	if err != nil {
		return -1, errors.Wrap(err, "failed to marshal order")
	}
	r, err := http.NewRequest("POST",
		cli.baseurl+"/order/certificate/"+productNameID,
		bytes.NewBuffer(orderBody))
	if err != nil {
		return -1, errors.Wrap(err, "failed to prepare order request")
	}
	resp, err := cli.Do(r)
	if err != nil {
		return -1, errors.Wrap(err, "failed to submit request to digicert api")
	}
	if resp.StatusCode != http.StatusCreated {
		return -1, errors.Errorf("failed to create order: %s", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, errors.Wrap(err, "failed to read response body")
	}
	defer resp.Body.Close()
	var or orderResponse
	err = json.Unmarshal(body, &or)
	if err != nil {
		return -1, errors.Wrap(err, "failed to parse JSON response body")
	}
	if len(or.Requests) != 1 {
		return -1, errors.New("no request id was found in digicert's response")
	}
	return or.Requests[0].ID, nil
}
