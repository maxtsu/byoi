package main

// Program calls the 3 components for the healthbot plugin
// Kafka module for ingesting data
// configjson for reading the /etc/byoi/config.json
// influxdb point write

import (
	"encoding/json"
	"fmt"
	"os"
	//"maxwell/configjson"
)

// the plugin config.json file
// var configfile = "/etc/byoi/config.json"
//var configfile = "config.json"

// TEST struct
type Person struct {
	Name string
	Age  int
}

func main() {
	fmt.Println("HelloWorld!")
	//configjson.ConfigJSON(configfile)
	//fmt.Println(configjson.Configuration.Hbin.Inputs[0].Plugin.Config.Device[1])

	//TESTING json writing
	file, err := os.Open("data.json")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	defer file.Close()
	var p Person
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&p)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	fmt.Println("persons name: " + p.Name)
}
