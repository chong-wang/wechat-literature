package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/songtianyi/rrframework/logs"
	"github.com/songtianyi/wechat-go/wxweb"
)

// register plugin
func RegisterReplier(session *wxweb.Session) {
	session.HandlerRegister.Add(wxweb.MSG_TEXT, wxweb.Handler(autoReply), "text-replier")
	session.HandlerRegister.Add(wxweb.MSG_LINK, wxweb.Handler(autoReply), "link-replier")
	if err := session.HandlerRegister.Add(wxweb.MSG_IMG, wxweb.Handler(autoReply), "img-replier"); err != nil {
		logs.Error(err)
	}

	if err := session.HandlerRegister.EnableByName("text-replier"); err != nil {
		logs.Error(err)
	}
	if err := session.HandlerRegister.EnableByName("link-replier"); err != nil {
		logs.Error(err)
	}
	if err := session.HandlerRegister.EnableByName("img-replier"); err != nil {
		logs.Error(err)
	}
}

func processListCmd(session *wxweb.Session, args []string, username string) {
	if len(args) == 0 {
		ls := ProgressListString()
		session.SendText(ls, session.Bot.UserName, username)
	} else {
		i, _ := strconv.ParseInt(args[0], 10, 64)
		ls := ProgressBookListString(int(i))
		session.SendText(ls, session.Bot.UserName, username)
	}
}

// "改书名 进度列表序号 书名列表序号 书名"
func processModifyBookName(session *wxweb.Session, args []string, username string) {
	if len(args) < 3 {
		return
	}

	who, _ := strconv.Atoi(args[0])
	book, _ := strconv.Atoi(args[1])
	name := args[2]

	nick, origin := ProgressModifyBookName(who, book, name)
	modify := fmt.Sprintf("改书名：%v 《%v》→《%v》", nick, origin, name)
	session.SendText(modify, session.Bot.UserName, username)
}

// "改进度 进度列表序号 书名列表序号 进度"
func processModifyBookPercent(session *wxweb.Session, args []string, username string) {
	if len(args) < 3 {
		return
	}

	who, _ := strconv.Atoi(args[0])
	book, _ := strconv.Atoi(args[1])
	percent, _ := strconv.Atoi(args[2])

	nick, name, origin := ProgressModifyBookPercent(who, book, percent)
	modify := fmt.Sprintf("改进度：%v 《%v》%v → %v", nick, name, origin, percent)
	session.SendText(modify, session.Bot.UserName, username)
}

// "删除进度 进度列表序号 书名列表序号"
func processDeleteBookProgress(session *wxweb.Session, args []string, username string) {
	if len(args) < 2 {
		return
	}

	who, _ := strconv.Atoi(args[0])
	book, _ := strconv.Atoi(args[1])
	nick, name, percent := ProgressDeleteBook(who, book)
	modify := fmt.Sprintf("删除进度：%v 《%v》%v", nick, name, percent)
	session.SendText(modify, session.Bot.UserName, username)
}

// 命令格式：“【命令名】 【操作对象】 【值】”
func processMyselfCommand(session *wxweb.Session, msg *wxweb.ReceivedMessage, username string) bool {
	if msg.MsgType != wxweb.MSG_TEXT {
		return false
	}
	fields := strings.Fields(msg.Content)
	if len(fields) == 0 {
		return false
	}
	switch fields[0] {
	case "命令":
		text := "改书名 进度列表序号 书名列表序号 书名\n" +
			"改进度 进度列表序号 书名列表序号 进度\n" +
			"删除进度 进度列表序号 书名列表序号\n" +
			"列表 [进度列表序号]\n" +
			"同步进度到文件\n" +
			"读取进度文件"
		session.SendText(text, session.Bot.UserName, username)
	case "列表":
		processListCmd(session, fields[1:], username)
	case "改书名":
		processModifyBookName(session, fields[1:], username)
	case "改进度":
		processModifyBookPercent(session, fields[1:], username)
	case "删除进度":
		processDeleteBookProgress(session, fields[1:], username)
	case "同步进度到文件":
		SyncProgress(DefaultProgressFile)
		session.SendText("Done", session.Bot.UserName, username)
	case "读取进度文件":
		LoadProgress(DefaultProgressFile)
		session.SendText("Done", session.Bot.UserName, username)
	default:
		return false
	}
	return true
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

	if processMyselfCommand(session, msg, username) {
		return
	}
	fillReportCache(session, username, nil)
	processTextMsg(session, msg, username)
}
