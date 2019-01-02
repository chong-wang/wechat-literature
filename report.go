package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/songtianyi/rrframework/logs"
	"github.com/songtianyi/wechat-go/wxweb"
)

const (
	LiteratureLoverGroup  = "360wenxueaihaozhe"
	LiteratureEditorGroup = "360wenxuebianjibu"
)

func RegisterReportProgress(session *wxweb.Session) {
	var types = []int{
		wxweb.MSG_TEXT,
		wxweb.MSG_LINK,
		wxweb.MSG_FV,
		wxweb.MSG_PF,
		wxweb.MSG_SCC,
		wxweb.MSG_INIT,
		wxweb.MSG_SYSNOTICE,
		wxweb.MSG_SYS,
	}

	for i, typ := range types {
		name := fmt.Sprintf("group-msg-%v", i)
		session.HandlerRegister.Add(typ, wxweb.Handler(reportProgress), name)
		if err := session.HandlerRegister.EnableByName(name); err != nil {
			logs.Error(err)
			os.Exit(1)
		}
	}
}

var reportCache struct {
	group  map[string]string
	mm     *wxweb.MemberManager
	update func()
}

func fillReportCache(session *wxweb.Session, username string, msg *wxweb.ReceivedMessage) {
	var once sync.Once
	reportCache.update = func() {
		once.Do(func() {
			updateMemberManager(session, username, msg)
		})
	}
}

func updateMemberManager(session *wxweb.Session, username string, msg *wxweb.ReceivedMessage) {
	contact := session.Cm.GetContactByUserName(username)
	if contact == nil {
		contact = &wxweb.User{UserName: username}
	}
	mm, err := wxweb.CreateMemberManagerFromGroupContact(session, contact)
	if err != nil {
		logs.Debug(err)
		return
	}
	if reportCache.group == nil {
		reportCache.group = make(map[string]string)
	}
	group := mm.Group.PYQuanPin
	reportCache.group[username] = group

	if group == LiteratureLoverGroup {
		reportCache.mm = mm
		checkMemberChanges(session, mm, msg)
	}
}

func checkMemberChanges(session *wxweb.Session, mm *wxweb.MemberManager, msg *wxweb.ReceivedMessage) {
	result := UpdateMembers(session, mm)
	if len(result.Joins) > 0 {
		if msg != nil && msg.MsgType == wxweb.MSG_SYS && len(result.Joins) < 3 {
			log.Printf("new join members: %#v, content: %v\n", result.Joins, msg.Content)
			for _, nick := range result.Joins {
				text := "欢迎新童鞋: @" + nick
				session.SendText(text, session.Bot.UserName, mm.Group.UserName)
			}
		}
	}
	if len(result.Leaves) > 0 {
		log.Printf("leave members: %#v\n", result.Leaves)
		for _, nick := range result.Leaves {
			MarkProgressLeave(nick)
		}
	}
	if len(result.ChangeNick) > 0 {
		for old, new := range result.ChangeNick {
			log.Printf("change nick: %q -> %q\n", old, new)
			ChangeProgressNick(old, new)
		}
	}

	if len(result.Joins)+len(result.Leaves)+len(result.ChangeNick) > 0 {
		SyncMembers(DefaultMemberFile)
	}
}

func getNick(session *wxweb.Session, who *wxweb.User) string {
	nick := GetNick(session, who)
	if !MemberExists(nick) {
		reportCache.update()
		if !MemberExists(nick) {
			logs.Error("member %q not exists", nick)
			return ""
		}
	}
	return nick
}

func processRobotKeywords(session *wxweb.Session, username string) {
	text := "机器人关键词：\n" +
		"进度：查看当前进度图片；\n" +
		"诗词：查看当天分享诗词；\n" +
		"借阅：查看文学社公共图书馆情况；\n" +
		"活动：查看当前线下活动报名情况（没有活动不回复）；\n" +
		"报名：报名当前线下分享活动（没有活动不回复）；\n" +
		"打卡：获取打卡系统链接: vip.lwhstudio.com, 如有问题请咨询Neo；\n" +
		"机器人、关键词：获取机器人关键词列表及说明。\n\n" +
		"机器人规则：\n" +
		"报进度格式：“书名 25%”；\n" +
		"分享诗词：发送古诗文网链接（例如https://m.gushiwen.org/shiwenv_0a4d69889c65.aspx）；\n" +
		"提醒报进度：每晚十点半提醒大家汇报读书进度。"
	session.SendText(text, session.Bot.UserName, username)
}

func processReportNoUpdate(session *wxweb.Session, username string, who *wxweb.User) {
	nick := getNick(session, who)
	if nick == "" {
		return
	}

	NoUpdateProgress(nick)
	SyncProgress(DefaultProgressFile)
}

func processReport(session *wxweb.Session, msg *wxweb.ReceivedMessage, username string, who *wxweb.User) {
	book, percent := ParseReportInfo(msg.Content)
	if book == "" {
		return
	}

	nick := getNick(session, who)
	if nick == "" {
		return
	}

	UpdateProgress(nick, book, percent)
	SyncProgress(DefaultProgressFile)

	// 减少打扰
	return

	// text := fmt.Sprintf("@%s [握手] 收到: 《%s》 %d%%。", nick, book, percent)
	// session.SendText(text, session.Bot.UserName, username)
	session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, username)
	text := ShareActivity()
	if text == "" {
		text = "欢迎大家体验读书打卡系统: vip.lwhstudio.com, 如有问题请咨询Neo。"
	}
	session.SendText(text, session.Bot.UserName, username)
}

func processJoinActivity(session *wxweb.Session, msg *wxweb.ReceivedMessage, username string, who *wxweb.User) {
	nick := getNick(session, who)
	if nick == "" {
		return
	}

	JoinActivity(nick)
	SyncActivity(DefaultActivityFile)
	text := ShareActivity()
	if text != "" {
		session.SendText(text, session.Bot.UserName, username)
	}
}

func processActivityInfo(session *wxweb.Session, msg *wxweb.ReceivedMessage, username string) {
	text := ShareActivity()
	if text != "" {
		session.SendText(text, session.Bot.UserName, username)
	}
}

func processTodayPoems(session *wxweb.Session, msg *wxweb.ReceivedMessage, username string) {
	poems := GetTodayPoems()
	if len(poems) == 0 {
		text := "今天还没有人分享诗词哦！分享方式：回复“https://m.gushiwen.org/shiwenv_0a4d69889c65.aspx” [呲牙]"
		session.SendText(text, session.Bot.UserName, username)
		return
	}

	ls := make([]string, len(poems))
	for i := range poems {
		ls[i] = poems[i].String()
	}

	session.SendText(strings.Join(ls, "\n\n"), session.Bot.UserName, username)
}

func processSharePoem(session *wxweb.Session, msg *wxweb.ReceivedMessage, username string, who *wxweb.User) {
	nick := getNick(session, who)
	if nick == "" {
		return
	}

	var url string = msg.Url
	if url == "" {
		idx := strings.Index(msg.Content, "http")
		if idx < 0 {
			return
		}
		url = msg.Content[idx:]

		if fields := strings.Fields(url); len(fields) > 1 {
			url = fields[0]
		}
	}

	poem, err := LoadPoemFromURL(url)
	if err != nil {
		logs.Error("load poem from url %q error: %v", url, err)
		return
	}

	AddPoem(poem)
	SyncPoems(DefaultPoemFile)
	processTodayPoems(session, msg, username)
}

func processShareBook(session *wxweb.Session, username string) {
	session.SendImgFromBytes(SharedBooksImage(), "share.jpg", session.Bot.UserName, username)
	text := "实际情况以石墨表格为准：https://shimo.im/sheet/CZ2RXKbgp8wkWklJ/"
	session.SendText(text, session.Bot.UserName, username)
}

func processTextMsg(session *wxweb.Session, msg *wxweb.ReceivedMessage, username string) {
	mm := reportCache.mm
	if mm == nil {
		reportCache.update()
		mm = reportCache.mm
		if mm == nil {
			logs.Error("get member manager failed")
			return
		}
	}

	who := mm.GetContactByUserName(msg.Who)
	if who == nil {
		logs.Error("get contact by username: %v failed", msg.Who)
		return
	}

	in := func(w string, ws ...string) bool {
		for _, s := range ws {
			if w == s {
				return true
			}
		}
		return false
	}

	if IsReport(msg.Content) {
		processReport(session, msg, username, who)
	} else if in(msg.Content, "无更", "今日无更") {
		processReportNoUpdate(session, username, who)
	} else if in(msg.Content, "进度") {
		session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, username)
	} else if in(msg.Content, "诗词") {
		processTodayPoems(session, msg, username)
	} else if strings.HasPrefix(msg.Content, "http") || strings.HasPrefix(msg.Url, "http") {
		processSharePoem(session, msg, username, who)
	} else if in(msg.Content, "借阅") {
		processShareBook(session, username)
	} else if in(msg.Content, "报名") {
		processJoinActivity(session, msg, username, who)
	} else if in(msg.Content, "活动") {
		processActivityInfo(session, msg, username)
	} else if in(msg.Content, "机器人", "关键词") {
		processRobotKeywords(session, username)
	} else if in(msg.Content, "打卡") {
		text := "欢迎大家体验读书打卡系统: vip.lwhstudio.com, 如有问题请咨询Neo。"
		session.SendText(text, session.Bot.UserName, username)
	}
}

func reportProgress(session *wxweb.Session, msg *wxweb.ReceivedMessage) {
	if !msg.IsGroup {
		return
	}

	var username = msg.FromUserName
	if msg.FromUserName == session.Bot.UserName {
		username = msg.ToUserName
	}

	fillReportCache(session, username, msg)

	group, ok := reportCache.group[username]
	if !ok {
		reportCache.update()
		group, ok = reportCache.group[username]
		if !ok {
			logs.Error("update member manager failed, username: %v", username)
		}
	}

	logs.Info("new msg from: %v", group)
	switch group {
	default:
		return
	case LiteratureEditorGroup:
		processMyselfCommand(session, msg, username)
		return
	case LiteratureLoverGroup:
		// do following
	}
	UpdateSessionUserName(session, username)

	if msg.MsgType == wxweb.MSG_TEXT {
		processTextMsg(session, msg, username)
	} else {
		p, _ := json.Marshal(msg)
		log.Printf("new msg: %s", p)
	}

	reportCache.update()
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

// 定时任务

var sessionUserName struct {
	session  *wxweb.Session
	username string
	sync.Mutex
}

func InitSessionUserName(session *wxweb.Session) {
	group := session.Cm.GetContactByPYQuanPin(LiteratureLoverGroup)
	if group == nil {
		log.Printf("InitSessionUserName: cannot find %v", LiteratureLoverGroup)
		return
	}
	UpdateSessionUserName(session, group.UserName)

	updateMemberManager(session, group.UserName, nil)
}

func UpdateSessionUserName(session *wxweb.Session, username string) {
	sessionUserName.Lock()
	defer sessionUserName.Unlock()
	sessionUserName.session = session
	sessionUserName.username = username
}

func NoticeReportProgress() {
	if !IsWorkday() {
		return
	}

	sessionUserName.Lock()
	defer sessionUserName.Unlock()

	session := sessionUserName.session
	username := sessionUserName.username

	if session != nil && username != "" {
		session.SendImgFromBytes(GenImage(), "progress.jpg", session.Bot.UserName, username)
		text := "@所有人 进度来一波"
		session.SendText(text, session.Bot.UserName, username)
	}
}
