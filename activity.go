package main

import (
	"strings"
	"time"
)

const DefaultActivityFile = "activity.json"

type Activity struct {
	Time string   `json:"time"`
	Doc  string   `json:"doc"`
	Join []string `json:"join"`
}

var (
	currentActivity Activity
)

func LoadActivity(file string) {
	LoadFromFile(&currentActivity, file)
}

func SyncActivity(file string) {
	SyncToFile(currentActivity, file)
}

func parseTime(t string) time.Time {
	p, _ := time.ParseInLocation("2006-01-02 15:04:05", t, time.Now().Location())
	return p
}

func JoinActivity(who string) {
	if time.Now().After(parseTime(currentActivity.Time)) {
		return
	}
	for _, nick := range currentActivity.Join {
		if nick == who {
			return
		}
	}
	currentActivity.Join = append(currentActivity.Join, who)
}

func ShareActivity() string {
	if time.Now().Before(parseTime(currentActivity.Time)) {
		return currentActivity.Doc + "。参与人有：" +
			strings.Join(currentActivity.Join, "、")
	}
	return ""
}
