package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token        string
	Spreadsheet  string
	GLAARGSource string
	roster       = make(map[string]Roster)
	rostLock     sync.RWMutex

	guildRoleMap = make(map[*discordgo.Guild]map[string]string)
	grmLock      sync.RWMutex
)

// Roster represents a discord user tied to their membership data, including desired roles.
type Roster struct {
	Callsign     string
	DesiredRoles []string
	Member       *discordgo.Member
	OM           *HamOperator
}

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&Spreadsheet, "s", "", "Google Spreadsheet ID")
	flag.StringVar(&GLAARGSource, "g", "", "GLAARG VE Data Source")
	flag.Parse()

	if Token == "" {
		fmt.Printf("Missing required '-t <discord_token>' argument")
		os.Exit(1)
	}
	if Spreadsheet == "" {
		fmt.Printf("Missing required '-s <google_spreadsheet_id>' argument")
		os.Exit(2)
	}

}

func main() {

	readSheet()
	dg, err := setupBot()
	if err != nil {
		os.Exit(1)
	}

	time.Sleep(7 * time.Second)
	enforceMemberships(dg)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

}
