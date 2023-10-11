package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"
)

type Config struct {
	ApiFetchInterval int `json:"api_fetch_interval"` // in seconds
	EnabledApiEndpoints []string `json:"enabled_api_endpoints"`
	PersistentDataFilePath string `json:"datafile"`
	Recipients []Recipient `json:"recipients"`
	SmtpConfiguration SmtpSettings `json:"smtp"`
	Template MailTemplateConfig `json:"template"`
}

func NewConfig() Config {
	// Initial config
	c := Config{
		ApiFetchInterval: 60 * 10, // every 10 minutes,
		EnabledApiEndpoints: []string{"bay", "bund"},
		PersistentDataFilePath: "data",
		Recipients: []Recipient{},
		SmtpConfiguration: SmtpSettings{
			From: "WID Notifier <from@example.org>",
			User: "from@example.org",
			Password: "SiEhAbEnMiChInSgEsIcHtGeFiLmTdAsDüRfEnSiEnIcHt",
			ServerHost: "example.org",
			ServerPort: 587},
		Template: MailTemplateConfig{
			SubjectTemplate: "",
			BodyTemplate: "",
		},
	}
	return c
}

func checkConfig(config Config) {
	if len(config.Recipients) < 1 {
		fmt.Println("ERROR\tConfiguration is incomplete.")
		panic(errors.New("no recipients are configured"))
	}
	for _, r := range config.Recipients {
		if !mailAddressIsValid(r.Address) {
			fmt.Println("ERROR\tConfiguration includes invalid data.")
			panic(errors.New("'" + r.Address + "' is not a valid e-mail address"))
		}
		if len(r.Filters) < 1 {
			fmt.Println("ERROR\tConfiguration is incomplete.")
			panic(errors.New("recipient " + r.Address + " has no filter defined - at least {'any': true/false} should be configured"))
		}
	}
	if !mailAddressIsValid(config.SmtpConfiguration.From) {
		fmt.Println("ERROR\tConfiguration includes invalid data.")
		panic(errors.New("'" + config.SmtpConfiguration.From + "' is not a valid e-mail address"))
	}
}

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

type DataStore struct {
	filepath string
	prettyJSON bool
	data any
	fileMode fs.FileMode
}

func (ds *DataStore) save() error {
	var err error
	var data []byte
	if ds.prettyJSON {
		data, err = json.MarshalIndent(ds.data, "", "  ")
	} else {
		data, err = json.Marshal(ds.data)
	}
	if err != nil {
		return err
	}
	err = os.WriteFile(ds.filepath, data, ds.fileMode)
	return err
}

func (ds *DataStore) load() error {
	data, err := os.ReadFile(ds.filepath)
	if err != nil {
		return err
	}
	switch ds.data.(type) {
	case Config:
		d, _ := ds.data.(Config);
		err = json.Unmarshal(data, &d)
		if err != nil {
			return err
		}
		ds.data = d
	case PersistentData:
		d, _ := ds.data.(PersistentData);
		err = json.Unmarshal(data, &d)
		if err != nil {
			return err
		}
		ds.data = d
	}
	return err
}

func (ds *DataStore) init() error {
	err := ds.load()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		} else {
			// Write initial data
			err = ds.save()
		}
	}
	return err
}

func NewDataStore(savePath string, data any, prettyJSON bool, fileMode fs.FileMode) DataStore {
	// Initial data
	ds := DataStore{}
	ds.filepath = savePath
	ds.data = data
	ds.prettyJSON = prettyJSON
	ds.fileMode = fileMode
	if err := ds.init(); err != nil {
		// We don't like that, we panic
		panic(err)
	}
	return ds
}
