package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

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

func mapMembers(s *discordgo.Session, g *discordgo.Guild, r *roster) {
	mem, err := s.GuildMembers(g.ID, "", 1000)
	if err != nil {
		fmt.Printf("Error reading guild members for %s: %s", g.Name, err)
	}

	for _, m := range mem {
		/*if m.Nick != "" {
			fmt.Printf("Member %s (%s) [%s] has Roles %v\n", m.User.Username, m.Nick, m.User.ID, m.Roles)
		} else {
			fmt.Printf("Member %s [%s] has Roles %v \n", m.User.Username, m.User.ID, m.Roles)
		} */
		//rostLock.Lock()
		e := r.getEntry(m.User.ID)
		//	if e == nil {
		//		e = new(Roster)
		//}
		e.Member = m
		//roster[m.User.ID] = e
		//rostLock.Unlock()
	}
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	grmLock.Lock()
	guildRoleMap[event.Guild] = make(map[string]string)
	grmLock.Unlock()

	fmt.Printf("Joined server %s (ID: %s).\n", event.Guild.Name, event.Guild.ID)
	fmt.Printf("Currently connected to %d servers.\n", len(guildRoleMap))
	//mapMembers(s, event.Guild)
	mapRoles(event.Guild)

	//fmt.Printf("GuildMap: %v", guildRoleMap)
}

//findRoleID finds a role id for role name per discord server
func findRoleID(rn string, g *discordgo.Guild) (string, error) {
	for _, r := range g.Roles {
		if r.Name == rn {
			return r.ID, nil
		}
	}
	return "", fmt.Errorf("Role %s not found on Guild %s", rn, g.Name)
}

// mapRoles populates the guild role map  with all roles in the guild.
func mapRoles(g *discordgo.Guild) {
	grmLock.Lock()
	defer grmLock.Unlock()
	for _, r := range g.Roles {
		//fmt.Printf("Role  ID %s is %s\n", r.ID, r.Name)
		guildRoleMap[g][r.Name] = r.ID
	}
}

func enforceMemberships(s *discordgo.Session, r *roster) {
	// For each guild...
	for g, rm := range guildRoleMap {
		r.Lock.RLock()
		defer r.Lock.RUnlock()
		// Look at each roster entry...
		for id, ros := range r.Map {
			// And process each desired role...
			for _, rn := range ros.DesiredRoles {
				hasRole := false
				//Loop through all roles this member has, and skip if they already have it
				for i := range ros.Member.Roles {
					if rm[rn] == ros.Member.Roles[i] {
						hasRole = true
						break
					}
				}
				if hasRole == true {
					//fmt.Printf("In Guild \"%s\" User \"%s\" already has role \"%s\" (Skipping Add)\n", g.Name, ros.Member.User.Username, rn)
				} else {
					fmt.Printf("In Guild \"%s\" User \"%s\" added to role \"%s\"\n", g.Name, ros.Member.User.Username, rn)
					err := s.GuildMemberRoleAdd(g.ID, id, rm[rn])
					if err != nil {
						fmt.Printf("Error adding user to roster: %v\n", err)
					}
				}
			}
			// Clear desired roles now that we've applied them
			r.Map[id].DesiredRoles = nil
		}
	}
}

func randomRadarQuote() string {
	quotes := []string{
		"These are the forms to get the forms to order more forms, sir.",
		"Here's a mover and a groover and it ain't by Herbert Hoover. It's for all you animals and music lovers.",
		"(Radar, seeing Klinger in pants) \"Don't I know your sister?\"",
		"Dear Mrs. Burns, I regret to inform that your husband has been seen out of uniform, and maybe you would like to know with who.",
		"Testing, 1,2,3,4,5,6,7,8 testing. A, B, C, D, E, F, G, H, I got a gal in Kalamazoo...",
		"I'm afraid he's doing some very important sleeping for the army right now.",
		"Why don't you sirs act like sirs, sir?",
		"Are you going to be a mother, sir?",
		"If I don't eat regularly, everything solid in my body turns to liquid.",
		"Oh, I am fine. Well, not really, I am closer to lousy than fine.",
		"Get away from me before I get physically emotional!",
		"What? He changed to psychiatry? That's crazy!",
		"Poetry, right? That's great how they can rhyme and be hot at the same time. ",
		"It's Mrs. Colonel, your wife, sir. ",
		"I've never seen you in your underneath before.",
		"If you want a drink, sir, -- compliments Henry Blake -- brandy, scotch, vodka. And for your convenience, all in the same bottle.",
		"As usual, I'm writing slowly because I know you can't read fast.",
		"Well, I guess that's a bear we all gotta cross.",
		"Testing, tes...1,2,3. Testing, 1, 2. Radar here, uh..there's nobody on the radio now except 'Seoul City' Sue so I figured I'd keep you entertained by reading you a letter from my mom. Here it goes. Dear Son, I got your lovely letter. You certainly asked a lot of questions. About the car, you may. About Jennifer next door, yes. About Eleanor Simon, she did once or twice but not too much. About your uncle Albert, uh no on drinking, yes on AA. About the dog Leon, three times in the bedroom, once under the washer, and twice on the cat. Testing, testing. About the cat, we don't have one anymore. About your cousin Ernie, he's in the...(explosion) Oh! Oh! Here we go again! Watch out!",
		"She kicked me and then she messed all my files from M to Zee and everything... And then she got mad.",
		"Listen, buddy, we're a hospital! How would you like it if we fired patients at you??",
		"I can't hear you! Boy, you've got the war on loud there!",
		"My father didn't have me til he was sixty-three. First time we played peek-a-boo he had a stroke.",
		"My bear went off!",
		"I've looked everywhere except the nurses's showers. Oh no, sir, I couldn't look in there - there might be naked female personnel showering with their clothes off!",
		"I don't think that this place is turning out to be that great an experience for me. I mean, I work under terrible pressure, and there's lots of death and destruction and stuff, but other than that I don't think I'm getting much out of it.",
		"When my Uncle Ed came home from World War I, his mother could tell from the look in his eyes that he hadn't been a good boy in France. She cried for three days. I just know when I get home, my mother's going to look at me and chuckle for a week.",
		"I'm the only one who's gonna leave this place younger than I was when I came in!",
		"Where were you originally born? I mean, as a child.",
		"I cleaned it for two hours! There was another mess under it!",
	}

	min := 0
	max := len(quotes) - 1
	rand.Seed(time.Now().UnixNano())
	q := rand.Intn(max-min) + min

	return quotes[q]

}
