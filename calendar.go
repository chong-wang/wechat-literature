package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var lawDate struct {
	holiday map[string]bool
	workday map[string]bool // time format: 2006-1-2
	sync.Mutex
}

func IsWorkday() bool {
	now := time.Now()
	today := now.Format("2006-1-2")

	lawDate.Lock()
	defer lawDate.Unlock()

	if IsWeekend(now) {
		return lawDate.workday[today]
	}
	return !lawDate.holiday[today]
}

func IsWeekend(now time.Time) bool {
	weekday := now.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func LoadLawDate() {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		err := loadLawDate()
		if err == nil {
			return
		}
	}
}

func loadLawDate() error {
	const API = "https://sp0.baidu.com/8aQDcjqpAAV3otqbppnN2DJv/api.php"

	form := make(url.Values)
	now := time.Now()
	form.Add("query", fmt.Sprintf("%v年%v月", now.Year(), int(now.Month())))
	form.Add("resource_id", "6018")
	form.Add("format", "json")

	uri := API + "?" + form.Encode()
	resp, err := http.Get(uri)
	if err != nil {
		log.Printf("LoadLawDate: get %q error: %v", uri, err)
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
		Data   []struct {
			Holiday struct {
				List []struct {
					Date   string `json:"date"`
					Status string `json:"status"`
				} `json:"list"`
			} `json:"holiday"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("LoadLawDate: get %q decode json error: %v", uri, err)
		return err
	}

	if len(result.Data) == 0 {
		log.Printf("LoadLawDate: get %q status %q, no data", uri, result.Status)
		return fmt.Errorf("no data")
	}

	if len(result.Data[0].Holiday.List) == 0 {
		log.Printf("LoadLawDate: no holiday this month")
		return nil
	}

	holiday := make(map[string]bool)
	workday := make(map[string]bool)
	for _, e := range result.Data[0].Holiday.List {
		if e.Status == "1" {
			holiday[e.Date] = true
		} else if e.Status == "2" {
			workday[e.Date] = true
		}
	}

	lawDate.Lock()
	lawDate.holiday = holiday
	lawDate.workday = workday
	lawDate.Unlock()
	return nil
}
