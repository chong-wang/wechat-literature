package main

import (
	"os"

	"github.com/songtianyi/rrframework/logs"
	"github.com/songtianyi/wechat-go/wxweb"
)

func ChDataDir(dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}
	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func main() {
	logs.SetLevel(logs.LevelWarning)

	session, err := wxweb.CreateSession(nil, nil, wxweb.BACKGROUND_MODE)
	if err != nil {
		logs.Error(err)
		return
	}

	ChDataDir("data")

	LoadMembers(DefaultMemberFile)
	LoadProgress(DefaultProgressFile)
	CheckAllProgress()
	ReCalcRank()
	LoadPoems(DefaultPoemFile)

	RegisterReportProgress(session)
	RegisterJoinGroup(session)
	RegisterReplier(session)
	RegisterAllTasks()
	RunTasks()

	const sessionFile = "session.json"
	LoadSession(sessionFile, session)
	session.SetAfterLogin(func() error {
		SaveSession(sessionFile, session)
		InitSessionUserName(session)
		return nil
	})

	println("https://login.weixin.qq.com/l/" + session.QrcodeUUID)

	for {
		if err := session.LoginAndServe(len(session.GetCookies()) != 0); err != nil {
			logs.Error("session exit, %s", err)
			logs.Info("trying re-login with cache")
			if err := session.LoginAndServe(true); err != nil {
				logs.Error("re-login error or session down, %s", err)
			}
			if session, err = wxweb.CreateSession(nil, session.HandlerRegister, wxweb.BACKGROUND_MODE); err != nil {
				logs.Error("create new sesion failed, %s", err)
				break
			}
			session.SetAfterLogin(func() error {
				SaveSession(sessionFile, session)
				InitSessionUserName(session)
				return nil
			})
		} else {
			logs.Info("closed by user")
			break
		}
	}
}
