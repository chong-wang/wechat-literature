package main

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

const DefaultPoemFile = "poems.json"

type Poem struct {
	Who     string
	Title   string
	Author  string
	Dynasty string
	Content string
}

func (p Poem) String() string {
	return fmt.Sprintf("%v\n%v: %v\n%v",
		p.Title, p.Author, p.Dynasty, p.Content)
}

type DatePoems struct {
	Date  time.Time
	Poems []Poem
}

var AllPoems []DatePoems

func AddPoem(poem Poem) {
	var ps *DatePoems
	now := time.Now()
	for i, p := range AllPoems {
		if IsSameDay(now, p.Date) {
			ps = &AllPoems[i]
			break
		}
	}
	if ps == nil {
		AllPoems = append(AllPoems, DatePoems{Date: now.Truncate(24 * time.Hour)})
		ps = &AllPoems[len(AllPoems)-1]
	}
	for _, p := range ps.Poems {
		if p.Title == poem.Title {
			p.Who = poem.Who
			return
		}
	}
	ps.Poems = append(ps.Poems, poem)
}

func SyncPoems(file string) error {
	return SyncToFile(AllPoems, file)
}

func LoadPoems(file string) error {
	return LoadFromFile(&AllPoems, file)
}

func GetTodayPoems() []Poem {
	now := time.Now()
	for _, p := range AllPoems {
		if IsSameDay(now, p.Date) {
			return p.Poems
		}
	}
	return nil
}

func poemSiteAvailable(host string) bool {
	return host == "so.gushiwen.org" || host == "m.gushiwen.org"
}

func parseIdFromURL(URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	if !poemSiteAvailable(u.Hostname()) {
		return ""
	}
	id := path.Base(u.Path)
	if i := strings.IndexByte(id, '_'); i >= 0 {
		id = id[i+1:]
	}
	id = id[:len(id)-len(path.Ext(id))]
	id = "#contson" + id
	return id
}

func LoadPoemFromURL(URL string) (Poem, error) {
	id := parseIdFromURL(URL)
	if id == "" {
		return Poem{}, fmt.Errorf("unexpected url: %q", URL)
	}

	doc, err := goquery.NewDocument(URL)
	if err != nil {
		return Poem{}, err
	}

	sel := doc.Find(id)
	title := sel.Prev().Prev().Text()
	author := sel.Prev().Text()
	var dynasty string
	rs := []rune(author)
	for i, r := range rs {
		if r == '：' || r == ':' {
			author = string(rs[:i])
			dynasty = string(rs[i+1:])
		}
	}

	rs = []rune(strings.TrimSpace(sel.Text()))
	if len(rs) == 0 {
		return Poem{}, fmt.Errorf("no content")
	}

	for i, r := range rs {
		if unicode.IsSpace(r) {
			rs[i] = '\n'
		}
	}

	content := string(rs)
	return Poem{Title: title, Author: author, Dynasty: dynasty, Content: content}, nil
}
