package main

// this is a comment
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// var configfile = "/etc/byoi/config.json"
var configfile = "config.json"

//var configfile = "users.json"

// The Top struct for the config.json
type Top struct {
	Logging string `json:"logging"`
	Hbin    string `json:"hbin"`
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
	// OutPlugin []OutPlugin `json:"plugin"`
}

// The Inputs struct for the config.json
type Inputs struct {
	Plugin string `json:"plugin"`
}

// The Plugin struct for the config.json
type Plugin struct {
	Name string `json:"name"`
	//Config []Config `json:"config"`
}

// Users struct which contains
// an array of users
type Users struct {
	Users []User `json:"users"`
}

// User struct which contains a name
// a type and a list of social links
type User struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Age    int      `json:"Age"`
	Social []Social `json:"social"`
}

// Social struct which contains
// list of links
type Social struct {
	// Facebook string `json:"facebook"`
	// Twitter  string `json:"twitter"`
}

func configJSON() {
	file, err := os.Open(configfile)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	defer file.Close()

	byteResult, _ := ioutil.ReadAll(file)

	//var res map[string]interface{}
	//json.Unmarshal([]byte(byteResult), &res)
	//fmt.Println(res)

	// we initialize our Users array
	var configuration Top
	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteResult, &configuration)
	fmt.Println(configuration)
}

func main() {
	//read config file
	configJSON()

}
