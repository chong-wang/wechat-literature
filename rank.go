package main

import (
	"sort"
)

type Rank struct {
	Key    string
	Name   string
	Target int // 读完多少本可升本级
}

var Ranks = []Rank{
	Rank{Key: "Iron1", Name: "顽强黑铁Ⅰ", Target: 0},
	Rank{Key: "Iron2", Name: "顽强黑铁Ⅱ", Target: 1},
	Rank{Key: "Iron3", Name: "顽强黑铁Ⅲ", Target: 2},
	Rank{Key: "Iron4", Name: "顽强黑铁Ⅳ", Target: 3},
	Rank{Key: "Copper1", Name: "倔强青铜Ⅰ", Target: 4},
	Rank{Key: "Copper2", Name: "倔强青铜Ⅱ", Target: 5},
	Rank{Key: "Copper3", Name: "倔强青铜Ⅲ", Target: 6},
	Rank{Key: "Copper4", Name: "倔强青铜Ⅳ", Target: 7},
	Rank{Key: "Silver1", Name: "秩序白银Ⅰ", Target: 8},
	Rank{Key: "Silver2", Name: "秩序白银Ⅱ", Target: 9},
	Rank{Key: "Silver3", Name: "秩序白银Ⅲ", Target: 10},
	Rank{Key: "Silver4", Name: "秩序白银Ⅳ", Target: 11},
	Rank{Key: "Gold1", Name: "荣耀黄金Ⅰ", Target: 12},
	Rank{Key: "Gold2", Name: "荣耀黄金Ⅱ", Target: 13},
	Rank{Key: "Gold3", Name: "荣耀黄金Ⅲ", Target: 14},
	Rank{Key: "Gold4", Name: "荣耀黄金Ⅳ", Target: 15},
	Rank{Key: "Platinum1", Name: "尊贵铂金Ⅰ", Target: 16},
	Rank{Key: "Platinum2", Name: "尊贵铂金Ⅱ", Target: 17},
	Rank{Key: "Platinum3", Name: "尊贵铂金Ⅲ", Target: 18},
	Rank{Key: "Platinum4", Name: "尊贵铂金Ⅳ", Target: 19},
}

func init() {
	sort.Slice(Ranks, func(i, j int) bool {
		return Ranks[i].Target < Ranks[j].Target
	})
}

func FindRank(target int) Rank {
	for i := len(Ranks) - 1; i >= 0; i-- {
		if target >= Ranks[i].Target {
			return Ranks[i]
		}
	}
	return Rank{Key: "Iron1", Name: "顽强黑铁Ⅰ", Target: 0}
}

func LoadRanks(file string) error {
	var ranks []Rank
	err := LoadFromFile(&ranks, file)
	if err == nil {
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].Target < ranks[j].Target
		})
		Ranks = ranks
	}
	return err
}

func SyncRanks(file string) error {
	return SyncToFile(Ranks, file)
}
