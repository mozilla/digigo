package digigo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// An Organization in the Digicert account
type Organization struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	AssumedName string `json:"assumed_name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	IsActive    bool   `json:"is_active,omitempty"`
	Address     string `json:"address"`
	Address2    string `json:"address2,omitempty"`
	Zip         string `json:"zip"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Telephone   string `json:"telephone,omitempty"`
	Container   struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		IsActive bool   `json:"is_active"`
	} `json:"container"`
	Validations []struct {
		Type           string    `json:"type"`
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		DateCreated    time.Time `json:"date_created,omitempty"`
		ValidatedUntil time.Time `json:"validated_until,omitempty"`
		Status         string    `json:"status"`
		VerifiedUsers  []struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"verified_users,omitempty"`
	} `json:"validations,omitempty"`
	EvApprovers []struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"ev_approvers,omitempty"`
}

type orgsList struct {
	Organizations []Organization `json:"organizations"`
}

// ListOrganizations returns the list of organizations for a given Digicert account
func (cli Client) ListOrganizations() ([]Organization, error) {
	r, err := http.NewRequest("GET", cli.baseurl+"/organization", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare request to list organizations")
	}
	resp, err := cli.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve list of organizations from digicert api")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	defer resp.Body.Close()
	var ol orgsList
	err = json.Unmarshal(body, &ol)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse JSON response body")
	}
	return ol.Organizations, nil
}
