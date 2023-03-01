package configjson

// Package accepts json config file
// Will extract to a struct

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// The Top struct for the config.json
type Top struct {
	Logging string `json:"logging"`
	Hbin    Hbin   `json:"hbin"`
	// Logging Logging `json:"logging"`
}

// The Logging struct for the config.json
type Logging struct {
}

// The Hbin struct for the config.json
type Hbin struct {
	Inputs  []Inputs  `json:"inputs"`
	Outputs []Outputs `json:"outputs"`
}

// The Outputs struct for the config.json
type Outputs struct {
	OutPlugin OutPlugin `json:"plugin"`
}

// The Output Plugin struct for the config.json
type OutPlugin struct {
	Name      string    `json:"name"`
	OutConfig OutConfig `json:"config"`
}

// The Config struct for the config.json
type OutConfig struct {
	Server   string `json:"server"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// The Inputs struct for the config.json
type Inputs struct {
	Plugin Plugin `json:"plugin"`
}

// The Plugin struct for the config.json
type Plugin struct {
	Name   string `json:"name"`
	Config Config `json:"config"`
}

// The Config struct for the config.json
type Config struct {
	Device []Device `json:"device"`
	Devgrp string   `json:"device-group"`
	KVs    []KVs    `json:"kvs"`
}

// Device type
type Device struct {
	Name      string         `json:"name"`
	SystemID  string         `json:"system-id"`
	Sensor    []Sensor       `json:"sensor"`
	HBStorage HBStorage      `json:"healthbot-storage"`
	Auth      Authentication `json:"authentication"`
}

// Sensor config
type Sensor struct {
	Name        string `json:"name"`
	KVs         []KVs  `json:"kvs"`
	Measurement string `json:"measurement"`
}

// HBStorage to specify DB name and retention policy name for this device
type HBStorage struct {
	DB              string `json:"database"`
	RetentionPolicy string `json:"retention-policy"`
}

// Authentication config
type Authentication struct {
	Password Password `json:"password"`
}

// Password config
type Password struct {
	Password string `json:"password"`
	username string `json:"username"`
}

// KVs config
type KVs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var Configuration Top

// Parse /etc/byoi/config.json file
// create as a struct
func ConfigJSON(configfile string) {

	file, err := os.Open(configfile)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	defer file.Close()

	byteResult, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteResult, &Configuration)
}
