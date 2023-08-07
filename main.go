package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gologme/log"
)

var configfile = "/etc/byoi/config.json"

// var configfile = "config.json"
var rulesfile = "rules.json"

func main() {

	log.EnableLevelsByNumber(5)
	log.EnableFormattedPrefix()

	//convert the config.json to a struct
	byteResult := gnfingest.ReadFile(configfile)
	var configuration gnfingest.JSONfile
	err := json.Unmarshal(byteResult, &configuration)
	if err != nil {
		fmt.Println("Unmarshall error", err)
	}

	// Set logging level From config.json
	loggingLevel := configuration.Logging.Level
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

	// extract the brokers topics and configuration from the configjson KVS
	kafkaCfg := gnfingest.KVS_parsing(configuration.Hbin.Inputs[0].Plugin.Config.KVS, []string{"brokers", "topics", "saslusername", "saslpassword", "saslmechanism", "securityprotocol"})

	//list of devices configuration from configjson
	bootstrapServers := kafkaCfg[0]

	// Generate unique uuid for kafka group-id
	//group := string(Create_uuid())
	//log.Debugf("Created Kafka group ID (UUID) %s", group)

	rand.Seed(time.Now().UnixNano())
	// String
	charset := "abcdefghijklmnopqrstuvwxyz"
	// Getting random character
	c := charset[rand.Intn(len(charset))]

	// Display the character
	fmt.Println("random string")
	fmt.Println(string(c))
	group := c
	/////

	devices := configuration.Hbin.Inputs[0].Plugin.Config.Device
	// config.json list of device key from values under sensor for searching messages
	keys := []string{"path", "rule-id"}
	device_keys := gnfingest.DeviceDetails(devices, keys)
	//fmt.Printf("Device-Keys %+v\n", device_keys)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Create kafka consumer configuration fro kafkaCfg
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"sasl.mechanisms":    kafkaCfg[4],
		"security.protocol":  kafkaCfg[5],
		"sasl.username":      kafkaCfg[2],
		"sasl.password":      kafkaCfg[3],
		"group.id":           group,
		"session.timeout.ms": 6000,
		// Start reading from the first message of each assigned
		// partition if there are no previously committed offsets
		// for this group.
		"auto.offset.reset": "earliest",
		// Whether or not we store offsets automatically.
		"enable.auto.offset.store": false,
	})
	if err != nil {
		log.Error("Failed to create consumer. ", err)
		os.Exit(1)
	}
	log.Info("Created Consumer. ", consumer)

	topics := []string{kafkaCfg[1]}
	err = consumer.SubscribeTopics(topics, nil)

	run := true
	//run := false
	for run {
		fmt.Printf("waiting for kafka message\n")
		time.Sleep(5 * time.Second)

	}

	// Processing sample message
	//var sample = "Sample4.json"
	//var sample = "Sample20.json"
	var sample = "interface-state.json"
	//var sample = "isis-1.json"
	byteResult = gnfingest.ReadFile(sample)
	// Unmarshal JSON message into struct
	var kafkaMessage gnfingest.Message
	err = json.Unmarshal(byteResult, &kafkaMessage)
	if err != nil {
		fmt.Println("Unmarshal error", err)
	}

	// Extract message source IP remove port number
	messageSource := strings.Split(kafkaMessage.Source, ":")[0]
	fmt.Println("source: %s", messageSource)
	// Extract message path remove index values []
	re := regexp.MustCompile("[[].*?[]]")
	messagePath := re.ReplaceAllString(kafkaMessage.Updates[0].Path, "")
	fmt.Println("path: %s", messagePath)
	// Start matching message to configured rules
	for _, d := range device_keys {
		if (d.DeviceName == messageSource) && (d.KVS_path == messagePath) {
			//extract rule-id
			rule_id := d.KVS_rule_id
			fmt.Printf("rule-id: %+v\n", rule_id)
			// Extract rule from rules.json
			for _, f1 := range rules[rule_id].Fields {
				for _, f2 := range f1.Path {
					fmt.Println("f2: %+v\n", f2)
				}
				//path := f1.Path
				//fields := kafkaMessage.Updates.Values.State
			}
		}
	}

	// parse to extract path & indexes from "Path:" value in message
	var result []string
	var k1 = make(map[string][]gnfingest.KVS)
	result, k1 = gnfingest.PathExtract(kafkaMessage.Updates[0].Path)
	fmt.Println("path list: %+v\n", result)
	fmt.Println("map of keys: %+v\n", k1)

	Test_json_map(kafkaMessage.Updates[0].Values)

	log.Info("this is info log")
	log.Println("some interesting logging message")
	log.Debug("debug log")
	log.Debugf("debug f %+v", k1)
	log.Error("error log")
	log.Infoln("info ln")
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

func Create_uuid() string {
	newUUID, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(newUUID)
}
