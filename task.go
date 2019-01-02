package main

import (
	"sort"
	"time"
)

type Task struct {
	When time.Time
	Do   func()
}

var AllTasks []*Task

// when format: "15:04:05"
// register before RunTasks
func AddTask(when string, do func()) error {
	now := time.Now()
	t, err := time.ParseInLocation("15:04:05", when, now.Location())
	if err != nil {
		return err
	}

	t = t.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
	AllTasks = append(AllTasks, &Task{When: t, Do: do})
	return nil
}

func RunTasks() {
	err := AddTask("23:59:59", func() {
		for _, t := range AllTasks {
			t.When = t.When.Add(24 * time.Hour)
		}
	})
	if err != nil {
		panic(err)
	}
	sort.Slice(AllTasks, func(i, j int) bool {
		return AllTasks[i].When.Before(AllTasks[j].When)
	})

	go func() {
		for {
			for _, t := range AllTasks {
				time.Sleep(t.When.Sub(time.Now()))
				if t.Do != nil {
					t.Do()
				}
			}
		}
	}()
}

func RegisterAllTasks() {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	// 每晚10点半提醒大家报进度
	check(AddTask("22:30:00", NoticeReportProgress))

	// 每天凌晨更新日历
	check(AddTask("01:00:00", LoadLawDate))

	load := func() {
		LoadActivity(DefaultActivityFile)
		LoadAllSharedBooks(DefaultSharedBookFile)
	}
	t := time.Now().Truncate(time.Hour)
	d := 10 * time.Minute
	n := int(24 * time.Hour / d)
	for i := 0; i < n; i++ {
		check(AddTask(t.Format("15:04:05"), load))
		t = t.Add(d)
	}
}
