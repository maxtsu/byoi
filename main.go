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

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gologme/log"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// Global Variables
var batchSize = 10       // Influx write batch size
var flushInterval = 2000 // Influx write flush intervale

var configfile = "config.json"

// var configfile = "/etc/byoi/config.json"
var rulesfile = "rules.json"

// Global variables 'Rules'
var Rules = make(map[string]gnfingest.RulesJSON)

// Getting Env details for TAND and group-id from ENV
var tand_host = (os.Getenv("TAND_HOST") + ".healthbot")
var tand_port = os.Getenv("TAND_PORT")
var group = (os.Getenv("CHANNEL'") + "-golang1")

func main() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	//convert the config.json to a struct
	byteResult := gnfingest.ReadFile(configfile)
	var configjson gnfingest.Configjson
	err := json.Unmarshal(byteResult, &configjson)
	if err != nil {
		fmt.Println("config.json Unmarshall error", err)
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
	log.Infoln("Logging configured as ", strings.ToUpper(loggingLevel), ". Set at level ", level)

	// config.json list of device key from values under sensor for searching messages
	keys := []string{"path", "rule-id", "prefix"} //list of keys/parameters to extract from the KVS section
	device_details := configjson.DeviceDetails(keys)

	fmt.Printf("\nPre-Device_details: %+v\n", device_details)

	//Create InfluxDB client
	influxClient := InfluxdbClient(tand_host, tand_port)
	log.Infof("Client create with client %+v\n", influxClient)
	fmt.Printf("\n\nCliewnt: %+v\n", influxClient)
	// Create Influx writeAPI for each database (source) from device_details list
	for _, d := range device_details {
		databas := d.Database
		wapi := influxClient.WriteAPI("my-org", databas)
		fmt.Printf("wapi: %+v\n", wapi)
		d.DeviceDetailsWriteAPI(influxClient)
	}

	//fmt.Printf("\nPost-Device_details: %+v\n", device_details)

	// extract the brokers topics and configuration from the configjson KVS
	kafkaCfg := gnfingest.KVS_parsing(configjson.Hbin.Inputs[0].Plugin.Config.KVS, []string{"brokers", "topics",
		"saslusername", "saslpassword", "saslmechanism", "securityprotocol"})

	// Create kafka consumer configuration for kafkaCfg
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaCfg["brokers"],
		"sasl.mechanisms":    kafkaCfg["saslmechanism"],
		"security.protocol":  kafkaCfg["securityprotocol"],
		"sasl.username":      kafkaCfg["saslusername"],
		"sasl.password":      kafkaCfg["saslpassword"],
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

	// Kafka consumer subscribes to topics
	topics := []string{kafkaCfg["topics"]}
	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		log.Error("Failed to subscribe to topics: ", err)
		os.Exit(1)
	}
	log.Infoln("Kafka consumer subscribed to topics: ", topics)

	// Load rules.json into struct
	byteResult = gnfingest.ReadFile(rulesfile)
	var r1 []gnfingest.RulesJSON
	err = json.Unmarshal(byteResult, &r1)
	if err != nil {
		log.Error("rules.json Unmarshall error ", err)
	}
	// create map of structs key=rule-id
	//var rules = make(map[string]gnfingest.RulesJSON)
	for _, r := range r1 {
		Rules[r.RuleID] = r
	}
	fmt.Printf("Rules from rules.json: %+v", Rules)

	//run := true
	run := false // for hometest
	for run {
		fmt.Printf("waiting for kafka message\n")
		time.Sleep(2 * time.Second)
		select {
		case sig := <-sigchan:
			log.Warnf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			// Poll the consumer for messages or events
			event := consumer.Poll(400)
			if event == nil {
				continue
			}
			switch e := event.(type) {
			case *kafka.Message:
				// Process the message received.
				//fmt.Printf("Got a kafka message\n")
				log.Debugf("%% Message on %s: %s\n", e.TopicPartition, safeSlice(string(e.Value), 100))
				kafkaMessage := string(e.Value)
				fmt.Printf("\nkafkaMessage: %s\n", kafkaMessage) //Message in single string
				// Start processing message
				ProcessKafkaMessage([]byte(kafkaMessage), device_details)
				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}
				// We can store the offsets of the messages manually or let
				// the library do it automatically based on the setting
				// enable.auto.offset.store. Once an offset is stored, the
				// library takes care of periodically committing it to the broker
				// if enable.auto.commit isn't set to false (the default is true).
				// By storing the offsets manually after completely processing
				// each message, we can ensure atleast once processing.
				_, err := consumer.StoreMessage(e)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%% Error storing offset after message %s:\n",
						e.TopicPartition)
				}
			case kafka.Error:
				// Errors are informational, the client will try to
				// automatically recover.
				log.Errorf("%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					log.Errorf("Kafka error. All brokers down ")
				}
			default:
				log.Errorf("Ignored %v\n", e)
			}
		}
	}
	// Run sample files in home lab
	hometest(device_details) //call hometest function
}

// Function to safely slice a string
func safeSlice(slice string, length int) string {
	if len(slice) < length {
		length = len(slice)
	}
	return slice[:length]
}

// Process the raw kafka message pointer to message (we do not change it)
func ProcessKafkaMessage(kafkaMessage []byte, devices_details []gnfingest.Device_Details) {
	// Kafka message convert to JSON struct
	message := gnfingest.Message{}
	err := json.Unmarshal([]byte(kafkaMessage), &message)
	if err != nil {
		log.Errorf("Kafka message Unmarshal error", err)
	}
	msgVerify := message.MessageVerify() //Verify the message struct schema
	if msgVerify != nil {                // Not valid openconfig message
		log.Infof("Ignore error JSON message %s\n", msgVerify)
	} else { // Valid openconfig message
		//Extract source IP and Path from message
		messageSource := message.MessageSource()
		messagePrefix := message.Tags.Prefix
		messagePath := message.Tags.Path
		log.Debugf("Message Source: %s Prefix: %s Path: %s\n", messageSource, messagePrefix, messagePath)
		messageMatchRule := false //Flag for match in message processing
		//Start matching message to configured rules in config.json (device_keys)
		for _, d := range devices_details {
			// Match Source-Prefix-Path
			if (d.DeviceName == messageSource) && (d.KVS_prefix == messagePrefix) && (d.KVS_path == messagePath) {
				messageMatchRule = true //flag message has matched a device in (device_keys) config.json

				//Rule-id for processing
				//rule_id := d.KVS_rule_id
				//log.Infof("Processing rule Rule-ID: %s Device: %s\n", rule_id, d.DeviceName)
				//rule := rules[rule_id] // extract the rule in rules.json
				//log.Infof("Process rule-id %s\n", rule.RuleID)

				ProcessJsonMessage(&message, kafkaMessage, &d)
			}
		}
		if !messageMatchRule {
			log.Debugln("Message no matching rule in config.json")
		}
	}
}

func ProcessJsonMessage(message *gnfingest.Message, kafkaMessage []byte, device_details *gnfingest.Device_Details) {
	//Rule-id for processing
	rule_id := device_details.KVS_rule_id
	log.Infof("Processing rule Rule-ID: %s Device: %s\n", rule_id, device_details.DeviceName)
	rule := Rules[rule_id] // extract the rule in rules.json
	// From message extract Tags as rawdata
	var rawTags gnfingest.MessageTags
	err := json.Unmarshal(kafkaMessage, &rawTags)
	if err != nil {
		log.Errorf("Message openconfig.MessageTags Unmarshal error", err)
	}
	// Convert rawTags into mapped output for JSON searching
	var tagMap map[string]any
	err1 := json.Unmarshal(rawTags.Tags, &tagMap)
	if err1 != nil {
		log.Errorf("Message Tags to map Unmarshal error", err)
	}

	// Declare fields map with interface
	var fields = make(map[string]interface{}) // fields & indexes from message

	//list of Indexes from rule.json
	for _, i := range rule.IndexValues {
		//Get indexes from Tags mapped value
		value := fmt.Sprintf("%v", tagMap[i]) //Convert all values into string
		fields[i] = value                     // Add value to data map
	}

	log.Debugf("Index fields from message: %+v\n", fields)
	// extract Field values
	getFields(message, fields, &rule)
	log.Debugf("Data & Index fields from message: %+v\n", fields)

	// data for InfluxDB
	WritePoint(fields, message, device_details)

}

func getFields(message *gnfingest.Message, fields map[string]interface{}, rule *gnfingest.RulesJSON) {
	// Receive raw data section of message (values) put in map
	var rawDataMap map[string]any
	err := json.Unmarshal(message.Values, &rawDataMap)
	if err != nil {
		log.Errorf("Kafka Values message Unmarshal error", err)
	}
	// list of fields to collect rule.Fields
	//Extract list of field values from the mapped values
	//var fields = make(map[string]string)
	for _, fi := range rule.Fields {
		fp := rule.Prefix + "/" + fi               //add prefix to path
		value := fmt.Sprintf("%v", rawDataMap[fp]) //Convert all values into string
		fields[fi] = value                         //Get field value Add to fields map
	}
	//log.Debugf("Data fields extracted from message: %+v\n", fields)
}

func hometest(device_keys []gnfingest.Device_Details) {
	// Processing sample message
	var sample = "ev-interface-state.json"
	//var sample = "isis-1.json"
	byteResult := gnfingest.ReadFile(sample)

	// fmt.Printf("message: %+v", kafkaMessage)
	//fmt.Printf("Device_keys: %+v", device_keys)
	// processs message just like a kafka message
	ProcessKafkaMessage(byteResult, device_keys)

}

// create InfluxDB client
func InfluxdbClient(tand_host string, tand_port string) influxdb2.Client {
	// set options for influx client
	options := influxdb2.DefaultOptions()
	options.SetBatchSize(uint(batchSize))
	options.SetFlushInterval(uint(flushInterval))
	options.SetLogLevel(2) //0 error, 1 - warning, 2 - info, 3 - debug

	// create client
	url := "https://" + tand_host + ":" + tand_port
	c := influxdb2.NewClientWithOptions(url, "my-token", options)
	defer c.Close()
	log.Infof("Created InfluxDB Client: %+v\n", c)
	return c //return the influx client
}

// Create the point with data for writing
func WritePoint(fields map[string]interface{}, msg *gnfingest.Message, dev *gnfingest.Device_Details) {
	tags := map[string]string{}
	time := time.Unix(msg.Timestamp, 0)
	p := influxdb2.NewPoint(dev.Measurement, tags, fields, time)
	fmt.Printf("Point: %+v\n", p)
	if dev.WriteApi != nil {
		//Write point to the writeAPI
		dev.WriteApi.WritePoint(p)
		log.Debugf("Write data point: %+v\n", p)
	} else {
		log.Errorf("WriteApi for: %+v <nil>\n", dev.DeviceName)
	}
}
