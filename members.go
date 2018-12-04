package main

import (
	"encoding/json"
	"os"

	"github.com/songtianyi/wechat-go/wxweb"
)

var Members []string

type MemberChanges struct {
	OldNewNick map[string]string
	Leaves     []string
	New        []string
}

func GetNick(session *wxweb.Session, who *wxweb.User) string {
	if who.DisplayName != "" {
		return who.DisplayName
	}
	friends := session.Cm.GetContactsByName(who.NickName)
	for _, friend := range friends {
		if friend.RemarkName == who.NickName {
			return friend.NickName
		}
	}
	return who.NickName
}

func getGroupMembers(session *wxweb.Session, list []*wxweb.User) []string {
	names := make([]string, len(list))
	for i, user := range list {
		names[i] = GetNick(session, user)
	}
	return names
}

func UpdateMembers(session *wxweb.Session, mm *wxweb.MemberManager) MemberChanges {
	// var i, j int
	// users := mm.Group.MemberList
	return MemberChanges{}
}

func FindMember(nick string) int {
	for i, mem := range Members {
		if nick == mem {
			return i
		}
	}
	return -1
}

func SyncMembers(file string) error {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	return enc.Encode(Members)
}

func LoadMembers(file string) error {
	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	return json.NewDecoder(fp).Decode(&Members)
}
