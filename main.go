package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gologme/log"
)

var configfile = "config.json"

// var configfile = "/etc/byoi/config.json"
var rulesfile = "rules.json"

func main() {
	// Getting Env details for TAND and group-id from ENV
	//tand_host := (os.Getenv("TAND_HOST") + ".healthbot")
	//tand_port := os.Getenv("TAND_PORT")
	//group := (os.Getenv("CHANNEL'") + "-golang1")

	//convert the config.json to a struct
	byteResult := gnfingest.ReadFile(configfile)
	var configjson gnfingest.Configjson
	err := json.Unmarshal(byteResult, &configjson)
	if err != nil {
		fmt.Println("Unmarshall error", err)
	}

	// Set logging level From config.json
	loggingLevel := configjson.Logging.Level
	var level int = 4
	switch loggingLevel {
	case "debug":
		level = 5
	case "info":
		level = 4
	case "warn":
		level = 3
	case "error":
		level = 2
	default:
		level = 4
	}
	// Initialize Logger
	// Level 10 = panic, fatal, error, warn, info, debug, & trace
	// Level 5 = panic, fatal, error, warn, info, & debug
	// Level 4 = panic, fatal, error, warn, & info
	// Level 3 = panic, fatal, error, & warn
	// Level 2 = panic, fatal & error
	// Level 1 = panic, fatal
	log.EnableLevelsByNumber(level)
	log.EnableFormattedPrefix()
	log.Info("Logging configured as ", strings.ToUpper(loggingLevel), ". Set at level ", level)

	// Load rules.json into struct
	byteResult = gnfingest.ReadFile(rulesfile)
	var r1 []gnfingest.RulesJSON
	err = json.Unmarshal(byteResult, &r1)
	if err != nil {
		log.Error("Unmarshall error ", err)
	}
	// create map of structs key=rule-id
	var rules = make(map[string]gnfingest.RulesJSON)
	for _, r := range r1 {
		rules[r.RuleID] = r
	}

	// config.json list of device key from values under sensor for searching messages
	devices := configjson.Hbin.Inputs[0].Plugin.Config.Device
	keys := []string{"path", "rule-id"}
	device_keys := gnfingest.DeviceDetails(devices, keys)
	fmt.Printf("Device-Keys %+v\n", device_keys)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	//run := false
	for run {
		fmt.Printf("waiting for kafka message\n")
		time.Sleep(2 * time.Second)
		select {
		case sig := <-sigchan:
			log.Warnf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			fmt.Println("Repeating message")
		}
	}

	// Run sample files in home lab
	//hometest()

}

// Process the raw kafka message pointer to message (we do not change it)
func ProcessKafkaMessage(message *gnfingest.Message, devices []gnfingest.Device_Keys) {
	//Extract source IP and Path from message
	messageSource := message.MessageSource()
	messagePath := message.MessagePath()
	log.Debugln("Source %s Path %s", messageSource, messagePath)

	//Start matching message to configured rules in config.json
	for _, d := range devices {
		fmt.Println("Device: ", d.DeviceName)
		if (d.DeviceName == message.Source) && (d.KVS_path == message.Updates[0].Path) {
			fmt.Printf("name and prefix match ")
		}
	}

}

func Test_json_map(rawdata json.RawMessage) {
	// Receive raw data section of message put in map
	var objMap map[string]any
	err := json.Unmarshal(rawdata, &objMap)
	if err != nil {
		fmt.Println("Unmarshal error", err)
	}
	fmt.Printf("Mapobject: %+v\n", objMap)
	// check for key (path)
	key := "interfaces/interface/state"
	value, ok := objMap[key]
	fmt.Printf("key %+v is there %+v\n", value, ok)
	if ok {
		// Unmarshall to correct struct
		var InterfaceState gnfingest.InterfacesInterfaceState
		err = json.Unmarshal(rawdata, &InterfaceState)
		if err != nil {
			fmt.Println("Unmarshal error", err)
		}
		fmt.Printf("\nstate struct: %+v\n", InterfaceState)

	}

}

func hometest() {
	// Processing sample message
	//var sample = "Sample4.json"
	//var sample = "Sample20.json"
	var sample = "interface-state.json"
	//var sample = "isis-1.json"
	byteResult := gnfingest.ReadFile(sample)
	// Unmarshal JSON message into struct
	var kafkaMessage gnfingest.Message
	err := json.Unmarshal(byteResult, &kafkaMessage)
	if err != nil {
		fmt.Println("Unmarshal error", err)
	}

	// parse to extract path & indexes from "Path:" value in message
	var result []string
	var k1 = make(map[string][]gnfingest.KVS)
	result, k1 = gnfingest.PathExtract(kafkaMessage.Updates[0].Path)
	fmt.Println("path list: %+v\n", result)
	fmt.Println("map of keys: %+v\n", k1)

	Test_json_map(kafkaMessage.Updates[0].Values)
}

func messageMatching(messageSource string, messagePath string, device_keys []gnfingest.Device_Keys) {
	// Start matching message to configured rules
	for _, d := range device_keys {
		if (d.DeviceName == messageSource) && (d.KVS_path == messagePath) {
			//extract rule-id
			rule_id := d.KVS_rule_id
			fmt.Printf("rule-id: %+v\n", rule_id)
			// Extract rule from rules.json
			//for _, f1 := range rules[rule_id].Fields {
			//	for _, f2 := range f1.Path {
			//		fmt.Println("f2: %+v\n", f2)
			//	}
			//path := f1.Path
			//fields := kafkaMessage.Updates.Values.State
		}
	}
}
