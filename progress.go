package main

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultProgressFile = "progress.json"

type Progress struct {
	Nick      string    `json:"nick" col:"昵称"`
	Books     []string  `json:"book" col:"书籍"`
	Percents  []int     `json:"percent" col:"完成度"`
	Rank      string    `json:"rank" col:"排位"`
	Blood     int       `json:"blood" col:"血条"`
	Completed int       `json:"completed" col:"本期完成书籍"`
	History   []string  `json:"history" col:"已完成"`
	UpdateAt  time.Time `json:"update_at"`
	Leaved    bool      `json:"leaved"`
}

var AllProgress []*Progress

func ProgressListString() string {
	ls := make([]string, 0, len(AllProgress))
	for i, p := range AllProgress {
		ls = append(ls, fmt.Sprintf("%v: %v", i, p.Nick))
	}
	return strings.Join(ls, "\n")
}

func ProgressBookListString(i int) string {
	if i < 0 || i >= len(AllProgress) {
		return ""
	}
	p := AllProgress[i]
	ls := make([]string, 0, len(p.Books))
	for i, b := range p.Books {
		ls = append(ls, fmt.Sprintf("%v: %v %v%%", i, b, p.Percents[i]))
	}
	return strings.Join(ls, "\n")
}

func ProgressModifyBookName(who int, book int, name string) (string, string) {
	if who < 0 || who >= len(AllProgress) {
		return "", ""
	}
	p := AllProgress[who]
	if book < 0 || book >= len(p.Books) {
		return p.Nick, ""
	}
	origin := p.Books[book]
	p.Books[book] = name
	return p.Nick, origin
}

func ProgressModifyBookPercent(who int, book int, percent int) (string, string, int) {
	if who < 0 || who >= len(AllProgress) {
		return "", "", 0
	}
	p := AllProgress[who]
	if book < 0 || book >= len(p.Books) {
		return p.Nick, "", 0
	}
	origin := p.Percents[book]
	p.Percents[book] = percent
	return p.Nick, p.Books[book], origin
}

func ProgressDeleteBook(who int, book int) (string, string, int) {
	if who < 0 || who >= len(AllProgress) {
		return "", "", 0
	}
	p := AllProgress[who]
	if book < 0 || book >= len(p.Books) {
		return p.Nick, "", 0
	}
	name := p.Books[book]
	percent := p.Percents[book]
	p.Books = append(p.Books[:book], p.Books[:book+1]...)
	p.Percents = append(p.Percents[:book], p.Percents[:book+1]...)
	return p.Nick, name, percent
}

func findByNick(nick string) *Progress {
	var p *Progress
	for i := range AllProgress {
		if AllProgress[i].Nick == nick {
			p = AllProgress[i]
			break
		}
	}
	return p
}

func NoUpdateProgress(nick string) {
	p := findByNick(nick)
	if p == nil {
		return
	}
	p.UpdateAt = time.Now()
}

func UpdateProgress(nick, book string, percent int) {
	if book == "" {
		return
	}
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	p := findByNick(nick)
	if p == nil {
		p = &Progress{Nick: nick, Rank: FindRank(0).Name, Blood: 3}
		AllProgress = append(AllProgress, p)
	}

	p.UpdateAt = time.Now()

	var i int
	for i = 0; i < len(p.Books); i++ {
		if strings.EqualFold(p.Books[i], book) {
			break
		}
	}

	if i < len(p.Books) {
		p.Percents[i] = percent
	} else {
		p.Books = append(p.Books, book)
		p.Percents = append(p.Percents, percent)
	}

	if percent == 100 {
		p.Books = append(p.Books[:i], p.Books[i+1:]...)
		p.Percents = append(p.Percents[:i], p.Percents[i+1:]...)

		found := false
		for _, b := range p.History {
			if b == book {
				found = true
				break
			}
		}
		if !found {
			p.History = append(p.History, book)
			p.Completed++
			p.Rank = FindRank(len(p.History)).Name
		}
	}
}

func MarkProgressLeave(nick string) {
	p := findByNick(nick)
	if p == nil {
		return
	}
	p.Leaved = true
}

func ChangeProgressNick(old, new string) {
	p := findByNick(old)
	if p == nil {
		return
	}
	p.Nick = new
}

func GenImage() []byte {
	records, colors := ToRows()
	cells, rect := CalcTable(records)
	img := image.NewRGBA(rect)
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
	DrawTable(img, cells)
	DrawRecords(img, cells, records, colors)
	return ImageBytes(img)
}

func init() {
	all := AllProgress
	AllProgress = []*Progress{{}}
	rows, _ := ToRows()
	if len(rows[0]) != len(rows[1]) {
		panic("please check `func ToRows()', the column number not equal")
	}
	AllProgress = all
}

func CheckAllProgress() {
	for _, p := range AllProgress {
		if len(p.Books) != len(p.Percents) {
			panic("please check `" + p.Nick + "` progress")
		}
	}
}

func IsSameDay(d1, d2 time.Time) bool {
	return d1.Year() == d2.Year() && d1.Month() == d2.Month() && d1.Day() == d2.Day()
}

func ToRows() ([][]string, []image.Image) {
	now := time.Now()

	rows := make([][]string, 1+len(AllProgress))
	colors := make([]image.Image, len(rows))
	rows[0] = append(rows[0], "昵称", "书籍", "完成度", "排位", "血条", "本期完成书籍", "已完成")
	colors[0] = Gray

	for i, e := range AllProgress {
		percents := make([]string, len(e.Percents))
		for j, p := range e.Percents {
			percents[j] = strconv.Itoa(p) + "%"
		}
		rows[i+1] = append(rows[i+1], e.Nick)
		rows[i+1] = append(rows[i+1], strings.Join(e.Books, "、"))
		rows[i+1] = append(rows[i+1], strings.Join(percents, "、"))
		rows[i+1] = append(rows[i+1], string(e.Rank))
		rows[i+1] = append(rows[i+1], strconv.Itoa(e.Blood))
		rows[i+1] = append(rows[i+1], strconv.Itoa(e.Completed))
		rows[i+1] = append(rows[i+1], strings.Join(e.History, "、"))

		if IsSameDay(now, e.UpdateAt) {
			colors[i+1] = Green
		}
	}

	return rows, colors
}

func ReCalcRank() {
	for _, p := range AllProgress {
		p.Rank = FindRank(len(p.History)).Name
	}
}

func SyncProgress(file string) error {
	return SyncToFile(AllProgress, file)
}

func LoadProgress(file string) error {
	return LoadFromFile(&AllProgress, file)
}

func ArchiveProgress(file string) error {
	src, err := os.Open(file)
	if err != nil {
		return err
	}
	defer src.Close()

	archive := file + "." + time.Now().Add(-24*time.Hour).Format("20060102")
	dst, err := os.OpenFile(archive, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
