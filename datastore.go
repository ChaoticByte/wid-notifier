// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"encoding/json"
	"io/fs"
	"os"
)

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
	if err != nil { return err }
	err = os.WriteFile(ds.filepath, data, ds.fileMode)
	return err
}

func (ds *DataStore) load() error {
	data, err := os.ReadFile(ds.filepath)
	if err != nil { return err }
	switch ds.data.(type) {
	case Config:
		d, _ := ds.data.(Config);
		err = json.Unmarshal(data, &d)
		if err != nil { return err }
		ds.data = d
	case PersistentData:
		d, _ := ds.data.(PersistentData);
		err = json.Unmarshal(data, &d)
		if err != nil { return err }
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
