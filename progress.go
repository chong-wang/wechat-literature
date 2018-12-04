package main

import (
	"encoding/json"
	"image"
	"image/draw"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultProgressFile = "progress.json"

type Rank string

const (
	Iron1     Rank = "顽强黑铁Ⅰ"
	Iron2     Rank = "顽强黑铁Ⅱ"
	Iron3     Rank = "顽强黑铁Ⅲ"
	Iron4     Rank = "顽强黑铁Ⅳ"
	Copper1   Rank = "倔强青铜Ⅰ"
	Copper2   Rank = "倔强青铜Ⅱ"
	Copper3   Rank = "倔强青铜Ⅲ"
	Copper4   Rank = "倔强青铜Ⅳ"
	Silver1   Rank = "秩序白银Ⅰ"
	Silver2   Rank = "秩序白银Ⅱ"
	Silver3   Rank = "秩序白银Ⅲ"
	Silver4   Rank = "秩序白银Ⅳ"
	Gold1     Rank = "荣耀黄金Ⅰ"
	Gold2     Rank = "荣耀黄金Ⅱ"
	Gold3     Rank = "荣耀黄金Ⅲ"
	Gold4     Rank = "荣耀黄金Ⅳ"
	Platinum1 Rank = "尊贵铂金Ⅰ"
	Platinum2 Rank = "尊贵铂金Ⅱ"
	Platinum3 Rank = "尊贵铂金Ⅲ"
	Platinum4 Rank = "尊贵铂金Ⅳ"
)

type Progress struct {
	Nick      string    `json:"nick" col:"昵称"`
	Books     []string  `json:"book" col:"书籍"`
	Percents  []int     `json:"percent" col:"完成度"`
	Rank      Rank      `json:"rank" col:"排位"`
	Blood     int       `json:"blood" col:"血条"`
	Completed int       `json:"completed" col:"本期完成书籍"`
	History   []string  `json:"history" col:"已完成"`
	UpdateAt  time.Time `json:"update_at"`
}

var All []*Progress

func findByNick(nick string) *Progress {
	var p *Progress
	for i := range All {
		if All[i].Nick == nick {
			p = All[i]
			break
		}
	}
	return p
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
		p = &Progress{Nick: nick, Rank: Iron1, Blood: 3}
		All = append(All, p)
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
	all := All
	All = []*Progress{{}}
	rows, _ := ToRows()
	if len(rows[0]) != len(rows[1]) {
		panic("please check `func ToRows() [][]string', the column number not equal")
	}
	All = all
}

func CheckAll() {
	for _, p := range All {
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

	rows := make([][]string, 1+len(All))
	colors := make([]image.Image, len(rows))
	rows[0] = append(rows[0], "昵称", "书籍", "完成度", "排位", "血条", "本期完成书籍", "已完成")
	colors[0] = Gray

	for i, e := range All {
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

func SyncProgress(file string) error {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	return enc.Encode(All)
}

func LoadProgress(file string) error {
	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	return json.NewDecoder(fp).Decode(&All)
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
