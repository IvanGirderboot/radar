package main

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
)

func setupBot() (*discordgo.Session, error) {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return nil, err
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
		return nil, fmt.Errorf("error opening connection: %v", err)
	}
	return dg, nil
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

	if strings.Contains(m.Content, "radar") {
		s.ChannelMessageSend(m.ChannelID, randomRadarQuote())
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

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	// Check to make sure we don't add an existing guild
	//   This can happen on disconnect/reconnects
	found := false
	for _, g := range guildsJoined {
		if event.Guild.ID == g.ID {
			found = true
			break
		}
	}
	if !found {
		guildsJoined = append(guildsJoined, event.Guild)
	}

	log.Info(fmt.Sprintf("Joined server %s (ID: %s).\n", event.Guild.Name, event.Guild.ID))
	log.Info(fmt.Sprintf("Currently connected to %d servers.\n", len(guildsJoined)))
}

// mapRoles returns a map of text role name to role id (both as strings)
func mapRoles(g *discordgo.Guild) map[string]string {
	rm := make(map[string]string)
	for _, r := range g.Roles {
		//fmt.Printf("Role  ID %s is %s\n", r.ID, r.Name)
		rm[r.Name] = r.ID
	}
	return rm
}

// enforceMemerships adds all desired roles in every guild for everyone in the roster.
func enforceMemberships(s *discordgo.Session, r *roster) {
	// Stats are fun!
	var rolesAdded int
	var rolesRemoved int

	// For each guild...
	for _, guild := range guildsJoined {
		log.Debug(fmt.Sprintf("Processing guild %v", guild.Name))
		//Map the roles for this guild
		roleMap := mapRoles(guild)

		members, err := s.GuildMembers(guild.ID, "", 1000)
		if err != nil {
			log.WithFields(log.Fields{
				"routine": "Discord Role Membership",
			}).Error(fmt.Sprintf("Error fetching members for guild %v: %v", guild.Name, err))
		}
		// Look at each guild member
		for _, member := range members {
			log.WithFields(log.Fields{
				"routine": "Discord Role Membership",
			}).Debug(fmt.Sprintf("Proccessing guild memeber %v", member.User.Username))
			// Find the roster entry
			re := r.getEntry(member.User.ID)

			// Record username and nickname (if set) while we're here.
			go re.addDiscordInfo(member, guild)

			log.WithFields(log.Fields{
				"Routine":  "Discord Role Membership",
				"Function": "enforceMemberships",
			}).Trace(fmt.Sprintf("User %v has Roles %v", member.User.Username, member.Roles))

			// And process each desired role...
			log.WithFields(log.Fields{
				"routine": "Discord Role Membership",
			}).Debug(fmt.Sprintf("%v should have roles: %v", member.User.Username, re.DesiredRoles))
			for _, roleName := range re.DesiredRoles {
				_, found := Find(member.Roles, roleMap[roleName])
				// Add the role if we DON'T have it.
				if !found {
					log.WithFields(log.Fields{
						"Guild": guild.Name,
						"User":  member.User.Username,
						"Role":  roleName,
					}).Info("Added role to member.")
					err := s.GuildMemberRoleAdd(guild.ID, member.User.ID, roleMap[roleName])
					if err != nil {
						log.Error(fmt.Sprintf("Error adding user to role: %v\n", err))
					}
					rolesAdded++
				}
			}

			// Then process undesired roles...
			log.WithFields(log.Fields{
				"routine": "Discord Role Membership",
			}).Debug(fmt.Sprintf("%v should not have roles: %v", member.User.Username, re.UndesiredRoles))
			for _, roleName := range re.UndesiredRoles {
				_, found := Find(member.Roles, roleMap[roleName])
				// Remove the role if we DO have it.
				if found {
					log.WithFields(log.Fields{
						"Guild": guild.Name,
						"User":  member.User.Username,
						"Role":  roleName,
					}).Info("Removed role from member.")
					err := s.GuildMemberRoleRemove(guild.ID, member.User.ID, roleMap[roleName])
					if err != nil {
						log.Errorf("Error removing user from role: %v\n", err)
					}
					rolesRemoved++
				}
			}
		}
	}
	log.WithFields(log.Fields{
		"Roles Added":   rolesAdded,
		"Roles Removed": rolesRemoved,
	}).Info("Completed Discord role enforcement.")
}
