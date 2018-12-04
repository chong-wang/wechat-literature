package main

import (
	"encoding/json"
	"fmt"

	"github.com/songtianyi/rrframework/logs"
	"github.com/songtianyi/wechat-go/wxweb"
)

// register plugin
func RegisterReplier(session *wxweb.Session) {
	session.HandlerRegister.Add(wxweb.MSG_TEXT, wxweb.Handler(autoReply), "text-replier")
	if err := session.HandlerRegister.Add(wxweb.MSG_IMG, wxweb.Handler(autoReply), "img-replier"); err != nil {
		logs.Error(err)
	}

	if err := session.HandlerRegister.EnableByName("text-replier"); err != nil {
		logs.Error(err)
	}

	if err := session.HandlerRegister.EnableByName("img-replier"); err != nil {
		logs.Error(err)
	}
}

func autoReply(session *wxweb.Session, msg *wxweb.ReceivedMessage) {
	var username string
	if msg.FromUserName == session.Bot.UserName {
		username = msg.ToUserName
	} else {
		username = msg.FromUserName
	}
	if username != session.Bot.UserName {
		return
	}

	if msg.MsgType != wxweb.MSG_TEXT {
		return
	}

	contact := session.Cm.GetContactByPYQuanPin(LiteratureLoverGroup)
	mm, err := wxweb.CreateMemberManagerFromGroupContact(session, contact)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot, _ := json.MarshalIndent(session.Bot, "", "  ")
	fmt.Println(string(bot))
	// cm, _ := json.MarshalIndent(session.Cm.GetContactByPYQuanPin(LiteratureLoverGroup), "", "  ")
	// fmt.Println(string(cm))
	mb, _ := json.MarshalIndent(mm, "", "  ")
	fmt.Println(string(mb))
	// wb, _ := json.MarshalIndent(who, "", "  ")
	// fmt.Println(string(wb))

	status := make(map[uint32][]string)
	for _, user := range mm.Group.MemberList {
		status[user.AttrStatus] = append(status[user.AttrStatus], user.NickName)
	}
	for a, ns := range status {
		fmt.Println(a, ns)
	}
	return

	if IsReport(msg.Content) {
		book, percent := ParseReportInfo(msg.Content)
		text := fmt.Sprintf("收到: 《%s》 %d%%。", book, percent)
		session.SendText(text, session.Bot.UserName, session.Bot.UserName)
	} else if msg.Content == "进度" {
		session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, session.Bot.UserName)
	} else {
		session.SendText("主人，我还在呢", session.Bot.UserName, session.Bot.UserName)
	}
}
