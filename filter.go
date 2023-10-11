package main

import (
	"strings"
)

type Filter struct {
	Any bool `json:"any"`
	TitleContains string `json:"title_contains"`
	Classification string `json:"classification"`
	MinBaseScore int `json:"min_basescore"`
	Status string `json:"status"`
	ProductsContain string `json:"products_contain"`
	NoPatch string `json:"no_patch"`
}

func (f Filter) filter(notices []WidNotice) []WidNotice {
	filteredNotices := []WidNotice{}
	for _, n := range notices {
		matches := []bool{}
		if f.Any {
			matches = append(matches, true)
		} else 
		{
			if f.TitleContains != "" {
				matches = append(matches, strings.Contains(n.Title, f.TitleContains))
			}
			if f.Classification != "" {
				matches = append(matches, f.Classification == n.Classification)
			}
			if f.MinBaseScore > 0 {
				matches = append(matches, f.MinBaseScore <= n.Basescore)
			}
			if f.Status != "" {
				matches = append(matches, f.Status == n.Status)
			}
			if f.ProductsContain != "" {
				matches = append(matches, len(n.ProductNames) > 0)
			}
			if f.NoPatch != "" {
				matches = append(matches, f.NoPatch == n.NoPatch)
			}
		}
		allMatch := len(matches) > 0
		for _, m := range matches {
			if !m {
				allMatch = false
				break
			}
		}
		if allMatch {
			filteredNotices = append(filteredNotices, n)
		}
	}
	return filteredNotices
}
