package main

import (
	"encoding/json"
	"os"

	"github.com/songtianyi/wechat-go/wxweb"
)

const DefaultMemberFile = "members.json"

var AllMembers []string

type MemberChanges struct {
	ChangeNick map[string]string
	Leaves     []string
	Joins      []string
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

func MemberExists(nick string) bool {
	for _, n := range AllMembers {
		if nick == n {
			return true
		}
	}
	return false
}

func getGroupMembers(session *wxweb.Session, list []*wxweb.User) []string {
	names := make([]string, len(list))
	for i, user := range list {
		names[i] = GetNick(session, user)
	}
	return names
}

func UpdateMembers(session *wxweb.Session, mm *wxweb.MemberManager) MemberChanges {
	newMembers := getGroupMembers(session, mm.Group.MemberList)
	changes := compare(AllMembers, newMembers)
	AllMembers = newMembers
	return changes
}

const (
	OpNone   = 0 // "none"   // 0
	OpChange = 1 // "change" // 1
	OpLeave  = 2 // "leave"  // 2
	OpJoin   = 3 // "join"   // 3
)

type change struct {
	fromi int
	fromj int
	count int
	op    int
}

// 群成员变化只可能是：
//   1. 新加入成员
//   2. 成员离开
//   3. 改名
// 在以下情况下，该算法会和实际情况不符：
//   1. 最后一个成员退群，马上有新成员加入
//   2. 某成员退群，紧挨着他的下一个成员马上改名
func compare(old, new []string) MemberChanges {
	changes := make([][]change, len(old)+1)
	for i := range changes {
		changes[i] = make([]change, len(new)+1)
	}
	for i := 1; i <= len(old); i++ {
		changes[i][0].fromi = i - 1
		changes[i][0].count = i
		changes[i][0].op = OpLeave
	}
	for i := 1; i <= len(new); i++ {
		changes[0][i].fromj = i - 1
		changes[0][i].count = i
		changes[0][i].op = OpJoin
	}

	for i := 1; i <= len(old); i++ {
		for j := 1; j <= len(new); j++ {
			if changes[i-1][j].count <= changes[i][j-1].count {
				changes[i][j].count = changes[i-1][j].count + 1
				changes[i][j].fromi, changes[i][j].fromj = i-1, j
				changes[i][j].op = OpLeave
			} else {
				changes[i][j].count = changes[i][j-1].count + 1
				changes[i][j].fromi, changes[i][j].fromj = i, j-1
				changes[i][j].op = OpJoin
			}
			if old[i-1] == new[j-1] {
				changes[i][j].count = changes[i-1][j-1].count
				changes[i][j].fromi, changes[i][j].fromj = i-1, j-1
				changes[i][j].op = OpNone
			} else if changes[i-1][j-1].count+1 < changes[i][j].count {
				changes[i][j].count = changes[i-1][j-1].count + 1
				changes[i][j].fromi, changes[i][j].fromj = i-1, j-1
				changes[i][j].op = OpChange // 改名
			}
		}
	}

	var r MemberChanges
	for i, j := len(old), len(new); i > 0 || j > 0; {
		switch changes[i][j].op {
		case OpNone:
			i = changes[i][j].fromi
			j = changes[i][j].fromj
		case OpLeave:
			i = changes[i][j].fromi
			r.Leaves = append([]string{old[i]}, r.Leaves...)
		case OpChange:
			i = changes[i][j].fromi
			j = changes[i][j].fromj
			if r.ChangeNick == nil {
				r.ChangeNick = make(map[string]string)
			}
			r.ChangeNick[old[i]] = new[j]
		case OpJoin:
			j = changes[i][j].fromj
			// TODO: 最好不要在这做，当最后一个人退群，又有新人进来时会判断成改名
			r.Joins = append([]string{new[j]}, r.Joins...)
		}
	}
	return r
}

func SyncMembers(file string) error {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	return enc.Encode(AllMembers)
}

func LoadMembers(file string) error {
	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	return json.NewDecoder(fp).Decode(&AllMembers)
}
