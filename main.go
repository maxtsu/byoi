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
	client "github.com/influxdata/influxdb1-client/v2"
)

// Global Variables
const batchSize = 10       // Influx write batch size
const flushInterval = 6000 // Influx write flush interval in ms
const configfile = "config.json"

// const configfile = "/etc/byoi/config.json"
var rulesfile = "rules.json"

var Rules = make(map[string]gnfingest.RulesJSON)

// Getting Env details for TAND and group-id from ENV
var tand_host = (os.Getenv("TAND_HOST") + ".healthbot")
var tand_port = os.Getenv("TAND_PORT")
var group = (os.Getenv("CHANNEL'") + "-golang1")

func main() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Set global variables in gnfingest package
	gnfingest.GlobalVariables(batchSize, flushInterval)
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
	device_details_slice := configjson.DeviceDetails(keys)

	//Create Map of devices key is SystemID for source searching
	var device_details = make(map[string]*gnfingest.Device_Details)
	for _, d := range device_details_slice {
		device_details[d.SystemID] = d
	}

	//Create InfluxDB client
	//InfluxClient := gnfingest.InfluxCreateClient(tand_host, tand_port)
	gnfingest.InfluxCreateClient(tand_host, tand_port)
	log.Infof("Influx client create %+v\n", gnfingest.InfluxClient)

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
	// create map of structs key=rule-id using global Rules variable
	for _, r := range r1 {
		Rules[r.RuleID] = r
	}
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

	// Ensures background processes finishes
	gnfingest.InfluxClient.Close()
}

// Function to safely slice a string
func safeSlice(slice string, length int) string {
	if len(slice) < length {
		length = len(slice)
	}
	return slice[:length]
}

// Process the raw kafka message pointer to message (we do not change it)
func ProcessKafkaMessage(kafkaMessage []byte, devices_details map[string]*gnfingest.Device_Details) {
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
		messagePrefixPath := messagePrefix + messagePath
		log.Debugf("Message Source: %s Prefix: %s Path: %s\n", messageSource, messagePrefix, messagePath)
		messageMatchRule := false //Flag for match in message processing
		//Start matching message to configured rules in config.json (device_keys)
		device, ok1 := devices_details[messageSource]
		if ok1 {
			fmt.Printf("Device matched is... %+v\n", device)
			sensor, ok2 := device.Sensor[messagePrefixPath]
			if ok2 {
				fmt.Printf("Sensor is... %+v\n", sensor)
				messageMatchRule = true       //flag message has matched a device in config.json
				rule_id := sensor.KVS_rule_id //Rule-id for processing
				log.Infof("Processing rule Rule-ID: %s Name: %s SystemID: %s\n", rule_id, device.DeviceName, device.SystemID)
				ProcessJsonMessage(&message, &kafkaMessage, device, sensor)
			}
		}
		if !messageMatchRule {
			log.Debugln("Message no matching rule in config.json")
		}
	}
}

func ProcessJsonMessage(message *gnfingest.Message, kafkaMessage *[]byte, device *gnfingest.Device_Details, sensor gnfingest.Sensor) {
	//Rule-id for processing
	rule_id := sensor.KVS_rule_id
	rule := Rules[rule_id] // extract the rule in rules.json
	// From message extract Tags as rawdata
	var rawTags gnfingest.MessageTags
	err := json.Unmarshal(*kafkaMessage, &rawTags)
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

	//Extract timestamp in Unix format to golang format
	t := time.Unix(0, message.Timestamp)
	log.Debugf("Index fields from message: %+v\n", fields)
	// extract Field values
	getFields(message, fields, &rule)
	log.Debugf("Data & Index fields from message: %+v\n", fields)

	// Create InfluxDB Point and append to slice/list in Device_details using the handler function
	//device.AddPoint(Influxclient, fields, time, sensor, batchSize)

	// Create a point and add to batch
	//tags := map[string]string{}
	//pt, err := client.NewPoint(sensor.Measurement, tags, fields, time)
	pt, err := client.NewPoint(sensor.Measurement, nil, fields, t)
	//pt := *p
	if err != nil {
		log.Errorf("Device %s Create point error: %s\n", device.DeviceName, err.Error())
	} else {
		device.Point <- *pt //send batchpoint via the channel to the goroutine for TimerHandler
	}
	for i := 1; i < 100; i++ {
		fmt.Printf("For loop %+v\n", i)
		fmt.Printf("Device.. %+v\n", device.Points)
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("Device.. %+v\n", device)
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

func hometest(device_details map[string]*gnfingest.Device_Details) {
	// Processing sample message
	var sample = "ev-interface-state.json"
	//var sample = "isis-1.json"
	byteResult := gnfingest.ReadFile(sample)

	// fmt.Printf("message: %+v", kafkaMessage)
	//fmt.Printf("Device_keys: %+v", device_keys)
	// processs message just like a kafka message
	ProcessKafkaMessage(byteResult, device_details)
	//fmt.Printf("All Devices %+v\n", device_details)

}
