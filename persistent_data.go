// Copyright (c) 2024 Julian Müller (ChaoticByte)

package main

import (
	"time"
)

type PersistentData struct {
	// {endpoint id 1: time last published, endpoint id 2: ..., ...}
	LastPublished map[string]time.Time `json:"last_published"`
}

func NewPersistentData(c Config) PersistentData {
	// Initial persistent data
	d := PersistentData{LastPublished: map[string]time.Time{}}
	for _, e := range apiEndpoints {
		d.LastPublished[e.Id] = time.Now().Add(-time.Hour * 24) // a day ago
	}
	return d
}
