package main

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
	fmt.Println(configjson.Configuration.Hbin.Inputs[0].Plugin.Config.Device[1])
}
