package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token string
)

var serversJoined []*discordgo.Guild

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	// If the message is "pong" reply with "Ping!"
	sm := strings.Split(m.Content, " ")
	if sm[0] == "class" {
		om, err := callsignLookup(sm[1])
		if err != nil {
			fmt.Println("error looking up callsign,", err)
			s.ChannelMessageSend(m.ChannelID, "Que?")
		} else {
			msg := fmt.Sprintf("%s has license class %s", sm[1], om.LicenseClass)
			s.ChannelMessageSend(m.ChannelID, msg)
		}
	}
}

func updateLicenseClassRoles(s *discordgo.Session, g *discordgo.Guild) {
	mem, err := s.GuildMembers(g.ID, "", 1000)
	if err != nil {
		fmt.Printf("Error reading guild members for %s: %s", g.Name, err)
	}

	for _, m := range mem {
		if m.Nick != "" {
			fmt.Printf("Member %s (%s) [%s] has Roles %v \n", m.User.Username, m.Nick, m.User.ID, m.Roles)
		} else {
			fmt.Printf("Member %s [%s] has Roles %v \n", m.User.Username, m.User.ID, m.Roles)
		}

	}
	//s.GuildMemberRoleAdd
	//roles, err := s.GuildRoles()

}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}
	serversJoined = append(serversJoined, event.Guild)
	fmt.Printf("Joined server %s (ID: %s).\n", event.Guild.Name, event.Guild.ID)
	fmt.Printf("Currently connected to %d servers.\n", len(serversJoined))

	/*
		for _, channel := range event.Guild.Channels {
			if channel.ID == event.Guild.ID {
				_, _ = s.ChannelMessageSend(channel.ID, "Airhorn is ready! Type !airhorn while in a voice channel to play a sound.")
				return
			}
		}*/

	updateLicenseClassRoles(s, event.Guild)
}
