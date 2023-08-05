package gnfingest

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// The Top struct for the config.json
type JSONfile struct {
	Hbin    Hbin `json:"hbin"`
	Logging struct {
		Level   string `json:"level"`
		Enabled string `json:"enabled"`
	} `json:"logging"`
}

// The Hbin struct for the config.json
type Hbin struct {
	Inputs  []Inputs  `json:"inputs"`
	Outputs []Outputs `json:"outputs"`
}

type Outputs struct {
	Plugin struct {
		Name   string `json:"name"`
		Config struct {
			Server   string `json:"server"`
			Port     int    `json:"port"`
			User     string `json:"user"`
			Password string `json:"password"`
		} `json:"config"`
	} `json:"plugin"`
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
	Name   string `json:"name"`
	Sensor []struct {
		Name        string `json:"name"`
		KVS         []KVS  `json:"kvs"`
		Measurement string `json:"measurement"`
	} `json:"sensor"`
	Authentication struct {
		Password struct {
			Password string `json:"password"`
			Username string `json:"username"`
		} `json:"password"`
	} `json:"authentication"`
	HealthbotStorage struct {
		Database        string `json:"database"`
		RetentionPolicy string `json:"retention-policy"`
	} `json:"healthbot-storage"`
	SystemID string `json:"system-id"`
}

// KVS struct
type KVS struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// struct defining device details ::-
//
//		-device name
//		-kvs path
//	 -kvs rule-id
//		-measurement
//		-database
type Device_Keys struct {
	DeviceName  string
	KVS_path    string
	KVS_rule_id string
	Measurement string
	Database    string
}

// Function to read text file return byteResult
func ReadFile(fileName string) []byte {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("File reading error", err)
		return []byte{}
	}
	byteResult, _ := io.ReadAll(file)
	file.Close()
	return byteResult
}

// Function to return list/slice of device details from config.json
func DeviceDetails(devices []Device, keys []string) []Device_Keys {
	// create slice of devices
	device_keys := []Device_Keys{}
	for _, d := range devices {
		//extract list/slice of structs for sensors
		sensors := d.Sensor
		for _, s := range sensors {
			// Create Device_Keys struct and append it to the slice.
			var dev Device_Keys
			dev.DeviceName = d.SystemID
			dev.Database = d.HealthbotStorage.Database
			dev.Measurement = s.Measurement
			//extract the KVS pairs from sensor container
			values := KVS_parsing(s.KVS, keys)
			dev.KVS_path = values[0]
			dev.KVS_rule_id = values[1]
			device_keys = append(device_keys, dev)
		}
	}
	return device_keys
}

// Function to extract KVS pairs by key names passed into function
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

// Function to extract update path leaf values key-value(from path)
// return list (slice) of the path in folders
// returns a map of {folder,{key1:value1,key2:value2},etc.}
func PathExtract(path string) ([]string, map[string][]KVS) {
	//split into folders by forward slash Ignore slash inside square brackets
	var result []string
	var keys = make(map[string][]KVS)
	var kvs = []KVS{}
	var kstart int
	length := len(path)
	inbracket := false
	i := 0
	for j, c := range path {
		if c == '[' {
			if path[j-1] != ']' {
				result = append(result, path[i:j])
				kstart = j + 1
				inbracket = true
			} else {
				kstart = j + 1
				inbracket = true
			}
		} else if c == ']' {
			split := strings.Split(path[kstart:j], "=")
			k := KVS{Key: split[0], Value: split[1]}
			kvs = append(kvs, k)
			inbracket = false
		} else if c == '/' && !inbracket {
			if path[j-1] != ']' {
				result = append(result, path[i:j])
			} else {
				// add all kvs pairs to the map item for that folder
				keys[result[len(result)-1]] = kvs
				kvs = nil
			}
			i = j + 1
		} else if j == (length - 1) {
			result = append(result, path[i:])
		}
	}
	return result, keys
}

type Points struct {
	Measurement          string               `json:"measurement"`
	Time                 int64                `json:"time"`
	Source               string               `json:"source"`
	InterfaceStateFields InterfaceStateFields `json:"fields"`
}

// Interface state fields
type InterfaceStateFields struct {
	Admin_status string `json:"admin-status"`
	Oper_status  string `json:"oper-status"`
}

type RulesJSON struct {
	Comment     string `json:"comment"`
	RuleID      string `json:"rule-id"`
	Path        string `json:"path"`
	Prefix      string `json:"prefix"`
	IndexValues []struct {
		Path  string `json:"path"`
		Index string `json:"index"`
	} `json:"index_values"`
	Fields []struct {
		Path  []string `json:"path"`
		Value []string `json:"value"`
	} `json:"fields"`
}
