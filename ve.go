package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	fmt.Printf("Loading GLAARG VEC Data...\n")
	resp, err := http.Get(GLAARGSource)
	if err != nil {
		fmt.Printf("Error loading GLAARG Data: %v\n", err)
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	//reader.Comma = ';'
	for {
		data, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error processing a VE Record, skipping.\n")
			continue
		}
		//fmt.Printf("GLAARG VE Loaded with %v fields loaded: %v \n", len(data), data)
		glaargData[strings.ToUpper(data[2])] = data
	}
	fmt.Printf("Loaded %d records from the GLAARG VEC datastore.\n", len(glaargData))
}
