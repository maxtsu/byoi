package main

// Test program for linux compile
import (
	"fmt"
	"maxwell/configjson"
)

func main() {
	fmt.Println("HelloWorld!")
	configjson.ConfigJSON("config.json")

	fmt.Println("hbDB: " + configjson.Configuration.Hbin.Inputs[0].Plugin.Config.Device[0].HBStorage.DB)
}
