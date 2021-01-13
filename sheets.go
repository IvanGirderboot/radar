package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/apex/log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func readSheet(r *roster, c chan string) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// HRV Roster
	spreadsheetID := Spreadsheet
	readRange := "[Club Member]!A2:K"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		log.WithFields(log.Fields{
			"Routine":     "Google Sheets",
			"Sheet ID":    spreadsheetID,
			"Sheet Range": readRange,
		}).Error("No sheet data found.")
	} else {
		for offset, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			name := fmt.Sprintf("%v", row[3])
			cs := fmt.Sprintf("%v", row[5])
			did := fmt.Sprintf("%v", row[0])
			dun := fmt.Sprintf("%v", row[1])
			dnn := fmt.Sprintf("%v", row[2])
			var extraRoles []string
			if len(row) >= 11 {
				extraRoles = strings.Split(fmt.Sprintf("%v", row[10]), ",")
			}
			if did == "" {
				log.WithFields(log.Fields{
					"Callsign": cs,
					"Name":     name,
				}).Warnf("No Discord ID in Google sheet for %s (%s), skipping record.", cs, name)
				continue
			}

			e := r.getEntry(did)
			e.Callsign = cs
			_, found := Find(e.DesiredRoles, "Club Member")

			if !found {
				e.DesiredRoles = append(e.DesiredRoles, "Club Member")
			}

			// Add each extra role, if needed.
			for i := range extraRoles {
				role := strings.TrimSpace(extraRoles[i])
				_, found := Find(e.DesiredRoles, role)

				if !found {
					e.DesiredRoles = append(e.DesiredRoles, role)
				}
			}

			nicknames := strings.Join(e.Nicknames, ", ")
			// If there is no discord data to write to the Sheet, move on
			if (e.Username == "" || e.Username == dun) && (len(e.Nicknames) == 0 || nicknames == dnn) {
				continue
			}

			// Determine row to update
			editRange := fmt.Sprintf("[Club Member]!B%d:C", offset+2)

			var vr sheets.ValueRange

			editData := []interface{}{
				fmt.Sprintf("%s#%s", e.Username, e.Discriminator),
				nicknames}

			vr.Values = append(vr.Values, editData)

			_, err := srv.Spreadsheets.Values.Update(spreadsheetID, editRange, &vr).ValueInputOption("RAW").Do()
			if err != nil {
				log.WithFields(log.Fields{
					"Routine":     "Google Sheets",
					"Sheet ID":    spreadsheetID,
					"Sheet Range": editRange,
					"Values":      vr.Values,
					"Error":       err,
				}).Error("Error updating Google Sheet.")
			}
		}
	}
	c <- "Google Sheet processing complete."
}
