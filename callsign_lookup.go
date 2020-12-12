package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// HamOperator represents a Ham Radio Operator
type HamOperator struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Callsign      string `json:"callsign"`
	LicenseClass  string `json:"license_class"`
	LicenseStatus string `json:"license_status"`
}

var myClient = &http.Client{Timeout: 2 * time.Second}

func getJSON(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func callsignLookup(cs string) (*HamOperator, error) {

	dbSvr := "https://exam.tools/api/uls/individual/"

	url := dbSvr + cs

	om := new(HamOperator)
	err := getJSON(url, om)
	if err != nil {
		return om, err
	}

	return om, nil
}
