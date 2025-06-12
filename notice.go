// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	// "encoding/json"
	"time"
)

type WidNotice struct {
	// obligatory
	Uuid string `json:"uuid"`
	Name string `json:"name"`
	Title string `json:"title"`
	Published time.Time `json:"published"`
	Classification string `json:"classification"`
	// optional fields (only fully supported by cert-bund)
	Basescore int `json:"basescore"` // -1 = unknown
	Status string `json:"status"` // "" = unknown
	ProductNames []string `json:"productNames"` // empty = unknown
	Cves []string `json:"cves"` // empty = unknown
	NoPatch string `json:"noPatch"` // "" = unknown
	// metadata
	ApiEndpointId string
	PortalUrl string
}

func noticeSliceContains(notices []*WidNotice, notice *WidNotice) bool {
	for _, x := range notices {
		if x.Uuid == notice.Uuid {
			return true
		}
	}
	return false
}
