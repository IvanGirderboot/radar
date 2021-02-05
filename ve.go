package main

import (
	"strings"

	"github.com/apex/log"
)

// getVEStatus checks a callsign for status with each queriable
//   VE source and returns true if accrediated with at least one VE.
//   Also returns a slice of accrediting VECs
func isVE(cs string) (isVE bool, roles []string) {
	//fmt.Printf("Looking up VE status for callsign %s\n", cs)
	isVE = false
	cs = strings.ToUpper(cs) // Let's stop case issues in their tracks!

	if isGLAARGVE(cs) {
		isVE = true
		roles = append(roles, "VE (GLAARG)")
	}

	// Add generic VE role if they are a VE of any flavor
	if isVE {
		roles = append(roles, "VE")
	}

	return isVE, roles
}

// GLAARGVE represents a GLAARG VE
type GLAARGVE struct {
	Callsign   string `json:"veCallSign"`
	Active     string `json:"veAccreditationActive"`
	ReturnCode string `json:"type"`
}

// isGLAARGVE checks for an active registration with GLAARG-VEC
func isGLAARGVE(cs string) bool {
	dbSvr := "http://glaarglookup.n1cck.com:5000/ve?veCallSign="

	url := dbSvr + cs

	ve := new(GLAARGVE)
	err := getJSON(url, ve)
	if err != nil {
		log.WithFields(log.Fields{
			"Routine":  "GLAARG VE Lookup",
			"Error":    err,
			"Callsign": cs,
		}).Error("Error looking up GLAARG VE Record.")
		return false
	}
	if ve.Active == "Y" {
		return true
	}
	if ve.ReturnCode == "404" {
		log.WithFields(log.Fields{
			"Routine":  "GLAARG VE Lookup",
			"Callsign": cs,
		}).Debug("Callsign not found in GLAARG VE Database")
	}
	return false
}

// veLookup checks each rosterEntry in a roster and applies the VE role(s) as required.
func veLookup(r *roster, complete chan string) {
	for _, re := range r.Map {
		if re.Callsign == "" {
			continue
		}

		isVE, veRoles := isVE(re.Callsign)

		if isVE {
			//	re.DesiredRoles = append(re.DesiredRoles, veRoles...)
			for _, role := range veRoles {
				_, found := Find(re.DesiredRoles, role)
				if !found {
					re.DesiredRoles = append(re.DesiredRoles, role)
				}
			}
		}
	}
	complete <- "VE lookups complete."
}
