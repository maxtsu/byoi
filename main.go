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

var configfile = "config.json"

// var configfile = "/etc/byoi/config.json"
var rulesfile = "rules.json"

// Global variables 'Rules'
var Rules = make(map[string]gnfingest.RulesJSON)

// Getting Env details for TAND and group-id from ENV
var tand_host = (os.Getenv("TAND_HOST") + ".healthbot")
var tand_port = os.Getenv("TAND_PORT")

// var tand_host = "192.168.1.1"
// var tand_port = "8080"
var group = (os.Getenv("CHANNEL'") + "-golang1")

func main() {
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
	log.Info("Logging configured as ", strings.ToUpper(loggingLevel), ". Set at level ", level)

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

	// extract the brokers topics and configuration from the configjson KVS
	kafkaCfg := gnfingest.KVS_parsing(configjson.Hbin.Inputs[0].Plugin.Config.KVS, []string{"brokers", "topics",
		"saslusername", "saslpassword", "saslmechanism", "securityprotocol"})

	// config.json list of device key from values under sensor for searching messages
	keys := []string{"path", "rule-id", "prefix"} //list of keys/parameters to extract from the KVS section
	device_keys := configjson.DeviceDetails(keys)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

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

	topics := []string{kafkaCfg["topics"]}
	err = consumer.SubscribeTopics(topics, nil)

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
				ProcessKafkaMessage([]byte(kafkaMessage), device_keys)
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
	hometest(device_keys) //call hometest function
}

// Function to safely slice a string
func safeSlice(slice string, length int) string {
	if len(slice) < length {
		length = len(slice)
	}
	return slice[:length]
}

// Process the raw kafka message pointer to message (we do not change it)
func ProcessKafkaMessage(kafkaMessage []byte, devices_keys []gnfingest.Device_Key) {
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
		for _, d := range devices_keys {
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

func ProcessJsonMessage(message *gnfingest.Message, kafkaMessage []byte, device_key *gnfingest.Device_Key) {
	//Rule-id for processing
	rule_id := device_key.KVS_rule_id
	log.Infof("Processing rule Rule-ID: %s Device: %s\n", rule_id, device_key.DeviceName)
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
	time := message.Timestamp
	//fields_data := fields
	//var data = map[string]string

	// Write to TSDB
	//gnfingest.WriteTSDB(data)
	// connect influxDB create Influx client return batchpoint
	batchPoint, client := influxdbClient(tand_host, tand_port, device_key.Database)

	influxPoint(batchPoint, client, device_key, time, fields)

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

func hometest(device_keys []gnfingest.Device_Key) {
	// Processing sample message
	var sample = "ev-interface-state.json"
	//var sample = "isis-1.json"
	byteResult := gnfingest.ReadFile(sample)

	// fmt.Printf("message: %+v", kafkaMessage)
	//fmt.Printf("Device_keys: %+v", device_keys)
	// processs message just like a kafka message
	ProcessKafkaMessage(byteResult, device_keys)

}

// create InfluxDB v1.8 client
func influxdbClient(tand_host string, tand_port string, database string) (client.BatchPoints, client.Client) {
	url := "https://" + tand_host + ":" + tand_port
	// create client
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: url,
	})
	if err != nil {
		log.Errorf("Error creating InfluxDB Client: ", err)
	}
	defer c.Close()
	log.Infof("Created InfluxDB Client: %+v", c)

	// Create a new point batch for database
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database: database,
		//Precision: "s",
	})
	fmt.Printf("Client create with BP %+v %+v", bp, c)
	return bp, c
}

// Point defines the fields that will be written to the database
// Measurement, Time, and Fields are required
// Precision can be specified if the time is in epoch format (integer).
// Valid values for Precision are n, u, ms, s, m, and h
/* type Point struct {
	Measurement string
	Tags        map[string]string
	Time        time.Time
	Fields      map[string]interface{}
	Precision   string
	Raw         string
} */

// create InfluxDB v1.8 point
func influxPoint(bp client.BatchPoints, c client.Client, device *gnfingest.Device_Key, timestamp int64, fields map[string]interface{}) {
	/*
		#Create json data package for TSDB
		data = {"measurement":matching_rule[4],"time":timestamp,"source":matching_rule[1],"fields":fields}
	*/
	// Create a point and add to batch
	//tags := map[string]string{"source": device.DeviceName, "measurement": device.Measurement, "time": timestamp}
	tags := map[string]string{}
	// fields := map[string]interface{}{"admin-status": "UP","oper-status":  "DOWN",}
	time := time.Unix(0, timestamp)
	pt, err := client.NewPoint(device.Measurement, tags, fields, time)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	bp.AddPoint(pt)

	// Write the batch test
	err = c.Write(bp)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	} else {
		fmt.Printf("BP Wrote %+v", bp)
	}
}
