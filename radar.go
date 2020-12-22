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

	//roster
	guildRoleMap = make(map[*discordgo.Guild]map[string]string)
	grmLock      sync.RWMutex
)

// Roster represents a discord user tied to their membership data, including desired roles.

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

	//	readSheet()
	dg, err := setupBot()
	if err != nil {
		os.Exit(1)
	}

	r := newRoster()
	//	grm := newGuildRoleMap()

	// Let Bot connect first
	time.Sleep(5 * time.Second)
	//enforceMemberships(dg)

	ticker := time.NewTicker(30 * time.Second)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sc:
			// Cleanly close down the Discord session.
			fmt.Println("Exit signal received, quiting.")
			dg.Close()
			os.Exit(0)
		case t := <-ticker.C:
			fmt.Println("Beginning refresh cycle at ", t)
			sheetsRead := make(chan bool)
			go readSheet(r, sheetsRead)

			discordMapped := make(chan bool)
			go updateDiscordMaps(dg, discordMapped, r)

			<-sheetsRead
			fmt.Println("Sheets have been read")
			<-discordMapped
			fmt.Println("Discord Mapped")
			enforceMemberships(dg, r)
		}
	}
}

func updateDiscordMaps(s *discordgo.Session, c chan bool, r *roster) {
	grmLock.RLock()
	for g := range guildRoleMap {
		grmLock.RUnlock()
		mapRoles(g)
		mapMembers(s, g, r)
		grmLock.RLock()
	}
	grmLock.RUnlock()
	c <- true
}
