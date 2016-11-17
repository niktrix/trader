package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Port string `json:"port"`

	Db struct {
		IP   string `json:"ip"`
		Name string `json:"name"`

		Port string `json:"port"`
	} `json:"db"`
}

func readConfig() (err error) {
	var (
		file    *os.File
		decoder *json.Decoder
	)
	file, err = os.Open("config.json")
	if err != nil {
		return err
	}
	decoder = json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		return err
	}
	return nil

}
