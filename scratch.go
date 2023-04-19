package main

import (
	"byoi/configjson"
	"encoding/json"
	"fmt"
)

// var configfile = "/etc/byoi/config.json"
var configfile = "config.json"

// Message struct
//{
//	"source": "node14:57400",
//	"subscription-name": "port_stats",
//	"timestamp": 1677602769296000000,
//	"time": "2023-02-28T16:46:09.296Z",
//	"prefix": "openconfig-interfaces:",
//	"updates": [

type Message struct {
	Source            string `json:"source"`
	Subscription_name string `json:"subscription-name"`
	Timestamp         int64  `json:"timestamp"`
	Time              string `json:"time"`
	Prefix            string `json:"prefix"`
	Updates           string `json:"updates"`
}

func main() {
	//convert the config.json to a struct
	configjson.ConfigJSON(configfile)

	// extract the brokers and topics from the configjson KVS
	brokertopic := configjson.KVS_parsing(configjson.Configuration.Hbin.Inputs[0].Plugin.Config.KVS, []string{"brokers", "topics"})
	fmt.Println("this is the broker: " + brokertopic[0])
	fmt.Println("this is the topics: " + brokertopic[1])

	devices := configjson.Configuration.Hbin.Inputs[0].Plugin.Config.Device
	device_array := configjson.GetDevices(devices)
	fmt.Println("array %s", device_array)

	bootstrapServers := brokertopic[0]
	topics := []string{brokertopic[1]}

	fmt.Println("bootstrap: %s", bootstrapServers)
	fmt.Println("topics: %s", topics)

	for _, d := range devices {
		fmt.Println("Device: ", d.SystemID)
		fmt.Println("Path: ", d.Sensor)
	}

}

type source_prefix struct {
	Source  string `json:"source"`
	Prefix  string `json:"prefix"`
	Updates []path `json:"updates"`
}

type path struct {
	Path string `json:"Path"`
}

func get_source_prefix(message string) source_prefix {
	m := source_prefix{}
	fmt.Printf("Message: %s", message)
	json.Unmarshal([]byte(message), &m)
	println("source: %s", m.Source)
	println("prefix: %s", m.Prefix)
	println("Updates: %s", m.Updates[0].Path)
	return m
}
