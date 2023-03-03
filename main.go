package main

// Program calls the 3 components for the healthbot plugin
// Kafka module for ingesting data
// configjson for reading the /etc/byoi/config.json
// influxdb point write

import (
	"fmt"
	"maxwell/configjson"
)

// the plugin config.json file
// var configfile = "/etc/byoi/config.json"
var configfile = "config.json"

func main() {
	fmt.Println("HelloWorld!")
	configjson.ConfigJSON(configfile)

	fmt.Println("hbDB: " + configjson.Configuration.Hbin.Inputs[0].Plugin.Config.Device[0].HBStorage.DB)
}
