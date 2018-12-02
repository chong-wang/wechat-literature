package main

import (
	"github.com/songtianyi/rrframework/logs"
	"github.com/songtianyi/wechat-go/wxweb"
)

func main() {
	logs.SetLevel(logs.LevelWarning)

	session, err := wxweb.CreateSession(nil, nil, wxweb.BACKGROUND_MODE)
	if err != nil {
		logs.Error(err)
		return
	}

	LoadProgress(DefaultProgressFile)

	RegisterReportProgress(session)
	RegisterJoinGroup(session)

	const sessionFile = "session.json"
	LoadSession(sessionFile, session)
	session.SetAfterLogin(func() error {
		SaveSession(sessionFile, session)
		return nil
	})

	for {
		if err := session.LoginAndServe(len(session.GetCookies()) != 0); err != nil {
			logs.Error("session exit, %s", err)
			logs.Info("trying re-login with cache")
			if err := session.LoginAndServe(true); err != nil {
				logs.Error("re-login error or session down, %s", err)
			}
			if session, err = wxweb.CreateSession(nil, session.HandlerRegister, wxweb.TERMINAL_MODE); err != nil {
				logs.Error("create new sesion failed, %s", err)
				break
			}
			session.SetAfterLogin(func() error {
				SaveSession(sessionFile, session)
				return nil
			})
		} else {
			logs.Info("closed by user")
			break
		}
	}
}
