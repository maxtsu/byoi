package main

import (
	"byoi/configjson"
	"byoi/greeting"
	"fmt"
)

var configfile = "config.json"

func main() {
	fmt.Println("HelloWorld!")
	fmt.Println(greeting.WelcomeText)

	//convert the config.json to a struct
	configjson.ConfigJSON(configfile)
	fmt.Println("finicxhed process config.json")

}
