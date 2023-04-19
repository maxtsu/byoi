package configjson

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
	KVS    []KVS    `json:"kvs"`
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
	KVS         []KVS  `json:"kvs"`
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

// KVS config
type KVS struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// struct defining device details
//
//	-SensorPath
//	-deviceName
//	-kvs_path
//	-measurement
//	-database
type Device_item struct {
	SensorPath  string
	DeviceName  string
	KVS_path    string
	Measurement string
	Database    string
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

// Function to return array of device details from config.json
func GetDevices(devices []Device) []Device_item {
	// create slice of devices
	device_items := []Device_item{}
	for _, a := range devices {
		//extract list/slice of structs for sensors
		keys := a.Sensor
		for _, b := range keys {
			// Create device_item struct and append it to the slice.
			var d Device_item
			d.DeviceName = a.SystemID
			d.Database = a.HBStorage.DB
			d.Measurement = b.Measurement
			//extract the KVS pairs from sensor
			keys_path := b.KVS
			sensor_path := KVS_parsing(keys_path, []string{"path"})
			d.SensorPath = sensor_path[0]
			device_items = append(device_items, d)
		}
	}
	l := len(device_items)
	var device_array [l]Device_item
	return device_array
}

func KVS_parsing(keys []KVS, keyString []string) []string {
	// Build a config map:
	confMap := map[string]string{}
	for _, v := range keys {
		confMap[v.Key] = v.Value
	}
	// Find values by key in the config map
	//Return the value of keys requested in same order
	valueString := []string{}
	for _, key := range keyString {
		if v, ok := confMap[key]; ok {
			value := v
			//fmt.Println("keystring ", key)
			//fmt.Println("value ", value)
			valueString = append(valueString, value)
		} else {
			fmt.Println(key, "not in config.json")
			valueString = append(valueString, "")
		}
	}
	return valueString
}
