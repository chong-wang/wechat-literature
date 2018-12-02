package main

import (
	"fmt"

	"github.com/songtianyi/rrframework/logs"
	"github.com/songtianyi/wechat-go/wxweb"
)

func RegisterJoinGroup(session *wxweb.Session) {
	session.HandlerRegister.Add(wxweb.MSG_TEXT, wxweb.Handler(joinGroup), "sys-join-group")
	if err := session.HandlerRegister.EnableByName("sys-join-group"); err != nil {
		logs.Error(err)
	}
}

func joinGroup(session *wxweb.Session, msg *wxweb.ReceivedMessage) {
	if !msg.IsGroup {
		return
	}

	if msg.MsgType == wxweb.MSG_SYSNOTICE ||
		msg.MsgType == wxweb.MSG_SYS {
		fmt.Printf("system msg: %#v\n", msg)
		// TODO: 发欢迎消息
	}
}
