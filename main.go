package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"

	"github.com/bwmarrin/discordgo"
)

var (
	// Token is a command line variable for the Discort secret token
	Token string

	// Spreadsheet is a commond line variable for the Google sheet containsing the membership roster
	Spreadsheet string

	// GLAARGSource is a command line variable pointing to the GLAARG VE Data
	GLAARGSource string

	// LogLevel is an optional command line flag for the loggin level.
	LogLevel string

	// Global slice of guilds we are joined to presently
	guildsJoined []*discordgo.Guild
)

type roster struct {
	Map  map[string]*rosterEntry
	Lock sync.RWMutex
}

type rosterEntry struct {
	Callsign       string
	Username       string
	Discriminator  string
	Nicknames      []string
	DesiredRoles   []string
	UndesiredRoles []string
	OM             *HamOperator
}

// addDiscordInfo adds discord Username#Discriminator and Guild Nickname to a roster entry.
func (re *rosterEntry) addDiscordInfo(m *discordgo.Member, g *discordgo.Guild) {
	// Nicknames are optional, so don't record if not set.
	if m.Nick != "" {
		nick := fmt.Sprintf("%s: %s", g.Name, m.Nick)

		// Check if we already have a Nick for this guild, update if so.
		found := false
		for i, nn := range re.Nicknames {
			if strings.HasPrefix(nn, g.Name) {
				re.Nicknames[i] = nick
				found = true
				break
			}
		}
		if !found { // No existing entry for this guild, so add it
			re.Nicknames = append(re.Nicknames, nick)
		}

	}
	re.Username = m.User.Username
	re.Discriminator = m.User.Discriminator // This is the #1234 part of the Username

	log.WithFields(log.Fields{
		"Routine":       "Roster Update",
		"Callsign":      re.Callsign,
		"Nicknames":     re.Nicknames,
		"Username":      re.Username,
		"Discriminator": re.Discriminator,
	}).Debug("Updated roster entry")
}

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&Spreadsheet, "s", "", "Google Spreadsheet ID")
	flag.StringVar(&GLAARGSource, "g", "", "GLAARG VE Data Source")
	flag.StringVar(&LogLevel, "l", "INFO", "Log Level: [Debug|Info|Warn|Error|Fatal]")
	flag.Parse()

	log.SetHandler(cli.New(os.Stdout))
	switch strings.ToUpper(LogLevel) {
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)
		log.WithFields(log.Fields{
			"LogLevel Requested": LogLevel,
		}).Error("Unrecognized log level.  Defaulting to INFO.")
	}

	if Token == "" {
		fmt.Printf("Missing required '-t <discord_token>' argument")
		os.Exit(1)
	}
	if Spreadsheet == "" {
		fmt.Printf("Missing required '-s <google_spreadsheet_id>' argument")
		os.Exit(2)
	}

}

func (r *roster) getEntry(did string) *rosterEntry {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	if r.Map[did] == nil {
		n := new(rosterEntry)
		r.Map[did] = n
	}
	return r.Map[did]
}

func newRoster() *roster {
	r := new(roster)
	r.Map = make(map[string]*rosterEntry)
	//r.Lock = make(sync.RWMutex)
	return r
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func dispatchHandler(dg *discordgo.Session) {

	r := newRoster()
	apply := make(chan string)
	init := make(chan string, 1)

	sheetTicker := time.NewTicker(16 * time.Minute)
	veTicker := time.NewTicker(1 * time.Hour)
	csTicker := time.NewTicker(3 * time.Hour)
	veDbTicker := time.NewTicker(13 * time.Hour)

	init <- "Startup"

	for {
		select {
		case <-init: // Startup or when /radar refresh is called
			log.Info("Beginning full initialization cycle.")
			go readSheet(r, init)
			<-init
			go rosterCallsignLookup(r, init)
			go veLookup(r, apply)
			<-init
		case <-sheetTicker.C:
			log.Info("Beginning Google Sheets refresh cycle.")
			go readSheet(r, apply)
		case <-veTicker.C:
			log.Info("Beginning VE Lookup refresh cycle.")
			go veLookup(r, apply)
		case <-veDbTicker.C:
			//log.Info("Beginning VE Database refresh cycle.")
			//go loadGLAARGData()
		case <-csTicker.C:
			log.Info("Beginning Callsign Lookup refresh cycle.")
			go rosterCallsignLookup(r, apply)
		case reason := <-apply:
			log.WithFields(log.Fields{
				"Reason": reason,
			}).Info("Beginning Discord role enforcement.")
			go enforceMemberships(dg, r)
		}
	}
}

func main() {
	dg, err := setupBot()
	if err != nil {
		os.Exit(1)
	}

	go dispatchHandler(dg)

	log.Info("Bot is now running.  Press CTRL-C to exit.")

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	log.Info("Exit signal received, quiting.")
	dg.Close()
	os.Exit(0)
}
