package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/apex/log"
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

// rosterCallsignLookup performs a callsign lookup on each roster entry
//   and assigns the correct class-based role.
func rosterCallsignLookup(r *roster, complete chan bool) {
	for _, re := range r.Map {
		if re.Callsign == "" {
			continue
		}
		log.WithFields(log.Fields{
			"routine": "Callsign Lookups",
		}).Debug(fmt.Sprintf("Looking up data for %v", re.Callsign))
		om, err := callsignLookup(re.Callsign)
		if err != nil {
			log.WithFields(log.Fields{
				"routine": "Callsign Lookups",
			}).Error(fmt.Sprintf("Error looking up data for %v: %v", re.Callsign, err))
			continue
		}
		log.WithFields(log.Fields{
			"routine": "Callsign Lookups",
		}).Debug(fmt.Sprintf("%v has a %v class license with a %v status.", re.Callsign, om.LicenseClass, om.LicenseStatus))

		if om.LicenseStatus == "Active" {
			var removeRoles []string
			switch om.LicenseClass {
			case "Technician":
				removeRoles = []string{"General", "Amateur Extra"}
			case "General":
				removeRoles = []string{"Technician", "Amateur Extra"}
			case "Amateur Extra":
				removeRoles = []string{"Technician", "General"}
			default:
				log.WithFields(log.Fields{
					"routine": "Callsign Lookups",
				}).Warn(fmt.Sprintf("Unknown licence class %s for %s", om.LicenseClass, om.Callsign))
			}

			// Add the correct class, if needed
			_, found := Find(re.DesiredRoles, om.LicenseClass)
			if !found {
				log.WithFields(log.Fields{
					"routine": "Callsign Lookups",
				}).Info(fmt.Sprintf("Adding license class role %s for %s\n", om.LicenseClass, om.Callsign))
				re.DesiredRoles = append(re.DesiredRoles, om.LicenseClass)
			}

			// Set other classes to be removed.
			for _, role := range removeRoles {
				_, found := Find(re.UndesiredRoles, role)
				if !found {
					log.WithFields(log.Fields{
						"routine": "Callsign Lookups",
					}).Info(fmt.Sprintf("Removing license class role %s for %s\n", role, om.Callsign))
					re.UndesiredRoles = append(re.UndesiredRoles, role)
				}
			}
		}
	}
	complete <- true
}
