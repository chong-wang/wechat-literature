package main

import (
	//	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/songtianyi/rrframework/logs"
	"github.com/songtianyi/wechat-go/wxweb"
)

const LiteratureLoverGroup = "360wenxueaihaozhe"

func RegisterReportProgress(session *wxweb.Session) {
	session.HandlerRegister.Add(wxweb.MSG_TEXT, wxweb.Handler(reportProgress), "group-msg")
	if err := session.HandlerRegister.EnableByName("group-msg"); err != nil {
		logs.Error(err)
	}
}

func getNick(session *wxweb.Session, who *wxweb.User) string {
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
		names[i] = getNick(session, user)
	}
	return names
}

func reportProgress(session *wxweb.Session, msg *wxweb.ReceivedMessage) {
	if !msg.IsGroup {
		return
	}
	var contact *wxweb.User
	var username string
	if msg.FromUserName == session.Bot.UserName {
		username = msg.ToUserName
	} else {
		username = msg.FromUserName
	}
	contact = session.Cm.GetContactByUserName(username)
	if contact == nil {
		contact = &wxweb.User{UserName: username}
	}
	mm, err := wxweb.CreateMemberManagerFromGroupContact(session, contact)
	if err != nil {
		logs.Debug(err)
		return
	}
	fmt.Printf("new msg from: %v\n", mm.Group.PYQuanPin)
	if mm.Group.PYQuanPin != LiteratureLoverGroup {
		return
	}
	who := mm.GetContactByUserName(msg.Who)
	if who == nil {
		who = session.Bot
	}

	if msg.MsgType != wxweb.MSG_TEXT {
		return
	}

	// p, _ := json.Marshal(getGroupMembers(session, mm.Group.MemberList))
	// fmt.Printf("%s\n", p)

	if IsReport(msg.Content) {
		book, percent := ParseReportInfo(msg.Content)
		if book == "" {
			return
		}

		nick := getNick(session, who)
		UpdateProgress(nick, book, percent)
		SyncProgress(DefaultProgressFile)

		text := fmt.Sprintf("@%s [握手] 收到: 《%s》 %d%%。", nick, book, percent)
		// session.SendText(text, session.Bot.UserName, session.Bot.UserName)
		// session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, session.Bot.UserName)
		session.SendText(text, session.Bot.UserName, username)
		session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, username)
	} else if msg.Content == "进度" {
		session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, session.Bot.UserName)
	}
}

func IsReport(s string) bool {
	s = strings.TrimSpace(s)
	if !strings.HasSuffix(s, "%") {
		return false
	}
	return len(s) < 64 // 书名一般没多长
}

func ParseReportInfo(s string) (string, int) {
	if !IsReport(s) {
		return "", 0
	}
	s = strings.TrimSpace(s)

	// trim '%'
	s = strings.TrimSuffix(s, "%")
	var i int
	for i = len(s) - 1; i >= 0; i-- {
		if !unicode.IsDigit(rune(s[i])) {
			break
		}
	}
	percent, _ := strconv.ParseInt(s[i+1:], 10, 64)

	s = strings.TrimSpace(s[:i+1])
	if s == "" {
		return "", int(percent)
	}
	fields := strings.Fields(s)
	book := fields[len(fields)-1]
	book = strings.TrimFunc(book, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
	return book, int(percent)
}
