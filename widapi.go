// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// known API endpoints
var apiEndpoints []ApiEndpoint = []ApiEndpoint{
	{
		Id: "bay",
		EndpointUrl: "https://wid.lsi.bayern.de/content/public/securityAdvisory",
		PortalUrl: "https://wid.lsi.bayern.de/portal/wid/securityadvisory",
	},
	{
		Id: "bund",
		EndpointUrl: "https://wid.cert-bund.de/content/public/securityAdvisory",
		PortalUrl: "https://wid.cert-bund.de/portal/wid/securityadvisory",
	},
}

const PUBLISHED_TIME_FORMAT = "2006-01-02T15:04:05.999-07:00"
const USER_AGENT = "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/118.0"

var defaultParams = []string{
	"size=1000", // max backlog
	"sort=published,desc",
	"aboFilter=false",
}

type ApiEndpoint struct {
	Id string
	EndpointUrl string
	PortalUrl string
}

func (e ApiEndpoint) getNotices(since time.Time) ([]WidNotice, time.Time, error) {
	// returns a slice of WidNotice and the 'published' field of the last notice, and the error (or nil)
	var notices []WidNotice = []WidNotice{}
	var err error
	params := defaultParams
	// params = append(params, "publishedFromFilter=" + publishedFrom.Format(PUBLISHED_FROM_FILTER_TIME_FORMAT))
	// ^ looks like the API is f***ed, 'publishedFromFilter=...' does only factor in the day (-2h because of the
	// timezone), not the time of the day - echte Deutsche Wertarbeit mal wieder am Start
	// -> we have to filter by hand (see below)
	url := e.EndpointUrl + "?" + strings.Join(params, "&")
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", USER_AGENT)
	client := http.Client{}
	res, err := client.Do(req)
	if err == nil {
		if res.StatusCode == 200 {
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return []WidNotice{}, since, err
			}
			var decodedData map[string]interface{}
			if err = json.Unmarshal(resBody, &decodedData); err != nil {
				return []WidNotice{}, since, err
			}
			notices = parseApiResponse(decodedData, e)
		} else {
			logger.error(fmt.Sprintf("Get \"%v\": %v\n", url, res.Status))
			return []WidNotice{}, since, err
		}
	} else {
		return []WidNotice{}, since, err
	}
	if len(notices) > 0 {
		// And here the filtering begins. yay -.-
		noticesFiltered := []WidNotice{}
		lastPublished := since
		for _, n := range notices {
			if n.Published.After(since) {
				noticesFiltered = append(noticesFiltered, n)
				// while we are at it, we can also find lastPublished
				if n.Published.After(lastPublished) {
					lastPublished = n.Published
				}
			}
		}
		return noticesFiltered, lastPublished, nil
	} else {
		return []WidNotice{}, since, nil
	}
}

func parseApiResponse(data map[string]interface{}, apiEndpoint ApiEndpoint) []WidNotice {
	var notices []WidNotice = []WidNotice{}
	for _, d := range data["content"].([]interface{}) {
		d := d.(map[string]interface{})
		notice := WidNotice{
			Uuid: d["uuid"].(string),
			Name: d["name"].(string),
			Title: d["title"].(string),
			Classification: d["classification"].(string),
			ApiEndpointId: apiEndpoint.Id,
		}
		published, err := time.Parse(PUBLISHED_TIME_FORMAT, d["published"].(string))
		if err != nil {
			logger.error(err)
		}
		notice.Published = published
		// optional fields
		if v, ok := d["basescore"]; ok {
			notice.Basescore = int(v.(float64))
		} else {
			notice.Basescore = -1
		}
		if v, ok := d["status"]; ok {
			notice.Status = v.(string)
		}
		if v, ok := d["productNames"]; ok {
			for _, n := range v.([]interface{}) {
				notice.ProductNames = append(notice.ProductNames, n.(string))
			}
		}
		if v, ok := d["cves"]; ok {
			for _, c := range v.([]interface{}) {
				notice.Cves = append(notice.Cves, c.(string))
			}
		}
		if v, ok := d["noPatch"]; ok {
			if v.(bool) {
				notice.NoPatch = "true"
			} else {
				notice.NoPatch = "false"
			}
		}
		// metadata
		notice.PortalUrl = apiEndpoint.PortalUrl + "?name=" + notice.Name
		notices = append(notices, notice)
	}
	return notices
}
