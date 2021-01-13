package main

import (
	"encoding/json"
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
func rosterCallsignLookup(r *roster, complete chan string) {
	for _, re := range r.Map {
		if re.Callsign == "" {
			continue
		}
		log.WithFields(log.Fields{
			"Routine":  "Callsign Lookups",
			"Callsign": re.Callsign,
		}).Debug("Looking up callsign data")
		om, err := callsignLookup(re.Callsign)
		if err != nil {
			log.WithFields(log.Fields{
				"Routine":  "Callsign Lookups",
				"Error":    err,
				"Callsign": re.Callsign,
			}).Error("Error looking up callsign data.")
			continue
		}
		log.WithFields(log.Fields{
			"routine":        "Callsign Lookups",
			"Callsign":       re.Callsign,
			"License Class":  om.LicenseClass,
			"License Status": om.LicenseClass,
		}).Debug("License lookup successful.")

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
					"Routine":       "Callsign Lookups",
					"Callsign":      om.Callsign,
					"License Class": om.LicenseClass,
				}).Warn("Unknown licence class")
			}

			// Add the correct class, if needed
			_, found := Find(re.DesiredRoles, om.LicenseClass)
			if !found {
				log.WithFields(log.Fields{
					"Routine":  "Callsign Lookups",
					"Callsign": om.Callsign,
					"Role":     om.LicenseClass,
				}).Info("Marking license class role for addition")
				re.DesiredRoles = append(re.DesiredRoles, om.LicenseClass)
			}

			// Set other classes to be removed.
			for _, role := range removeRoles {
				_, found := Find(re.UndesiredRoles, role)
				if !found {
					log.WithFields(log.Fields{
						"Routine":  "Callsign Lookups",
						"Callsign": om.Callsign,
						"Role":     role,
					}).Info("Marking license class role for removal.")
					re.UndesiredRoles = append(re.UndesiredRoles, role)
				}
			}
		}
	}
	complete <- "Callsign lookups complete"
}
