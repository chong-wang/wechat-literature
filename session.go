package main

import (
	"encoding/json"
	"net/http"
	"os"
	"reflect"

	"github.com/songtianyi/wechat-go/wxweb"
)

type Session struct {
	WxWebCommon *wxweb.Common
	WxWebXcg    *wxweb.XmlConfig
	Cookies     []*http.Cookie
	Bot         *wxweb.User
	QrcodePath  string //qrcode path
	QrcodeUUID  string //uuid
	CreateTime  int64
	LastMsgID   string
}

func LoadSession(file string, session *wxweb.Session) {
	fp, err := os.Open(file)
	if err != nil {
		return
	}
	defer fp.Close()

	var sess Session
	err = json.NewDecoder(fp).Decode(&sess)
	if err != nil {
		return
	}

	src := reflect.ValueOf(sess)
	dst := reflect.ValueOf(session).Elem()
	typ := src.Type()
	for i := 0; i < typ.NumField(); i++ {
		dst.FieldByName(typ.Field(i).Name).Set(src.Field(i))
	}
}

func SaveSession(file string, session *wxweb.Session) {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer fp.Close()

	var sess Session
	src := reflect.ValueOf(session).Elem()
	dst := reflect.ValueOf(&sess).Elem()
	typ := dst.Type()
	for i := 0; i < typ.NumField(); i++ {
		dst.Field(i).Set(src.FieldByName(typ.Field(i).Name))
	}

	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	enc.Encode(sess)
}
