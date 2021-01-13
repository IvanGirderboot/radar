package main

import (
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/apex/log"
)

var (
	glaargData = make(map[string][]string)
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

// isGLAARGVE checks for an active registration with GLAARG-VEC
func isGLAARGVE(cs string) bool {
	if len(glaargData) < 1 {
		loadGLAARGData()
	}
	if _, ok := glaargData[cs]; ok {
		return true
	}
	return false
}

// loadGLAARG loads the GLAARG VE database into memory as there is no per-user lookup right now
func loadGLAARGData() {
	log.Info("Loading GLAARG VEC Data...")
	resp, err := http.Get(GLAARGSource)
	if err != nil {
		log.WithFields(log.Fields{
			"Routine": "VE Database",
			"Error":   err,
		}).Error("Error loading GLAARG Data")
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	//reader.Comma = ';'
	for {
		data, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.WithFields(log.Fields{
				"Routine": "VE Database",
				"Error":   err,
			}).Info("Error processing a VE Record, skipping.")
			continue
		}
		//fmt.Printf("GLAARG VE Loaded with %v fields loaded: %v \n", len(data), data)
		glaargData[strings.ToUpper(data[2])] = data
	}
	log.WithFields(log.Fields{
		"Routine":        "VE Database",
		"Records Loaded": len(glaargData),
	}).Info("Loaded the GLAARG VEC datastore.")
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
