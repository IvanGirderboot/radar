package main

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type roster struct {
	Map  map[string]*rosterEntry
	Lock sync.RWMutex
}

type rosterEntry struct {
	Callsign     string
	DesiredRoles []string
	Member       *discordgo.Member
	OM           *HamOperator
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

/*
type guildRoleMap struct {
	Map  map[*discordgo.Guild]map[string]string
	Lock sync.RWMutex
}

func newGuildRoleMap() *guildRoleMap {
	grm := new(guildRoleMap)
	grm.Map = make(map[*discordgo.Guild]map[string]string)
	return &grm
}

/*
func (grm *guildRoleMap) getGuild(gid *discordgo.Guild) {
	grm.Lock.Lock()
	defer grm.Lock.Unlock()
	if grm.Map[gid] = nil {
		grm.Map[gid] = make(map[string]string)
	}

}
*/
