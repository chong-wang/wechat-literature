package main

import (
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

var reportCache struct {
	group map[string]string
	mm    *wxweb.MemberManager
}

func reportProgress(session *wxweb.Session, msg *wxweb.ReceivedMessage) {
	if !msg.IsGroup {
		return
	}

	if msg.MsgType != wxweb.MSG_TEXT {
		return
	}

	var username = msg.FromUserName
	if msg.FromUserName == session.Bot.UserName {
		username = msg.ToUserName
	}

	var mm *wxweb.MemberManager
	group, ok := reportCache.group[username]
	if !ok {
		contact := session.Cm.GetContactByUserName(username)
		if contact == nil {
			contact = &wxweb.User{UserName: username}
		}
		var err error
		mm, err = wxweb.CreateMemberManagerFromGroupContact(session, contact)
		if err != nil {
			logs.Debug(err)
			return
		}
		if reportCache.group == nil {
			reportCache.group = make(map[string]string)
		}
		group = mm.Group.PYQuanPin
		reportCache.group[username] = group
	}

	fmt.Println("new msg from:", group)
	if group != LiteratureLoverGroup {
		return
	}

	if mm == nil {
		mm = reportCache.mm
	} else {
		reportCache.mm = mm
	}

	who := mm.GetContactByUserName(msg.Who)
	if who == nil {
		mm.Update(session)
		who = mm.GetContactByUserName(msg.Who)
	}
	if who == nil {
		who = session.Bot
	}

	if IsReport(msg.Content) {
		book, percent := ParseReportInfo(msg.Content)
		if book == "" {
			return
		}

		nick := GetNick(session, who)
		UpdateProgress(nick, book, percent)
		SyncProgress(DefaultProgressFile)

		// text := fmt.Sprintf("@%s [握手] 收到: 《%s》 %d%%。", nick, book, percent)
		// session.SendText(text, session.Bot.UserName, username)
		session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, username)
	} else if msg.Content == "进度" {
		session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, username)
	}
}

func IsReport(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	rs := []rune(s)
	r := rs[len(rs)-1]
	return len(rs) <= 20 && (r == '%' || r == '％')
}

func ParseReportInfo(s string) (string, int) {
	if !IsReport(s) {
		return "", 0
	}
	s = strings.TrimSpace(s)

	// trim '%'
	s = strings.TrimRightFunc(s, func(r rune) bool {
		return r == '%' || r == '％'
	})
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
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	return book, int(percent)
}
