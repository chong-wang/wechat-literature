package main

import (
	"encoding/json"
	"os"
)

func SyncToFile(v interface{}, file string) error {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func LoadFromFile(v interface{}, file string) error {
	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	return json.NewDecoder(fp).Decode(&v)
}
