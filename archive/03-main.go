package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gologme/log"
	client "github.com/influxdata/influxdb1-client/v2"
	"gopkg.in/yaml.v2"
)

// Global Variables
const batchSize = 600      // Influx write batch size
const flushInterval = 1000 // Influx write flush interval in ms
const configfile = "/etc/byoi/config.json"

var rulesfile = "/etc/healthbot/rules.yaml"
var Rules = make(map[string]gnfingest.YamlRule)

var Global_device_details = make(map[string]*gnfingest.Device_Details)

// Getting Env details for TAND and group-id from ENV
var tand_host = (os.Getenv("TAND_HOST") + ".healthbot")
var tand_port = os.Getenv("TAND_PORT")
var group = (os.Getenv("CHANNEL'") + "-golang1")

func main() {
	// signal channel to notify on SIGHUP
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP)

	// Set global variables in gnfingest package
	gnfingest.GlobalVariables(batchSize, flushInterval)
	//convert the config.json to a struct
	byteResult := gnfingest.ReadFile(configfile)
	var configjson gnfingest.Configjson
	err := json.Unmarshal(byteResult, &configjson)
	if err != nil {
		log.Errorln("config.json Unmarshall error", err)
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

	// Open log file and create if non-existent
	/*
		file, err := os.OpenFile("ingest_app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		// assign it to the standard logger and file
		multi := io.MultiWriter(file, os.Stdout)
		log.SetOutput(multi)  */

	// config.json list of device key from values under sensor for searching messages
	keys := []string{"path", "rule-id", "prefix"} //list of keys/parameters to extract from the KVS section

	device_details_slice := configjson.DeviceDetails(keys)
	//Create Map of devices key is SystemID for source searching
	for _, d := range device_details_slice {
		Global_device_details[d.SystemID] = d //Add to global variable
	}

	//Create InfluxDB client
	gnfingest.InfluxCreateClient(tand_host, tand_port)
	log.Infof("Influx client create %+v\n", gnfingest.InfluxClient)

	// extract the brokers topics and configuration from the configjson KVS
	kafkaCfg := gnfingest.KVS_parsing(configjson.Hbin.Inputs[0].Plugin.Config.KVS, []string{"brokers", "topics",
		"saslusername", "saslpassword", "saslmechanism", "securityprotocol"})
	ssl_ca_location := "/ca.crt" //localtion of local ca.crt certificate
	// Create kafka consumer configuration for kafkaCfg
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaCfg["brokers"],
		"sasl.mechanisms":    kafkaCfg["saslmechanism"],
		"security.protocol":  kafkaCfg["securityprotocol"],
		"sasl.username":      kafkaCfg["saslusername"],
		"sasl.password":      kafkaCfg["saslpassword"],
		"ssl.ca.location":    ssl_ca_location,
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

	// Check for event_rules.yaml file in healthbot folder
	if _, err := os.Stat(rulesfile); err == nil {
		log.Infof("rules.yaml file exists in healthbot folder\n")
	} else {
		rulesfile = "rules.yaml"
		log.Infof("rules.yaml file does not exists in healthbot folder. Using local held file\n")
	}
	// Load rules.yaml into struct
	byteResult = gnfingest.ReadFile(rulesfile)
	var r1 []gnfingest.YamlRule
	err = yaml.Unmarshal(byteResult, &r1)
	if err != nil {
		log.Error("rules.yaml Unmarshall error ", err)
	}
	// create map of structs key=rule-id using global Rules variable
	for _, r := range r1 {
		Rules[r.RuleID] = r
	}
	run := true
	for run {
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
				log.Debugf("%% Message on %s: %s\n", e.TopicPartition, safeSlice(string(e.Value), 400))
				kafkaMessage := string(e.Value)
				//fmt.Debugf("\nkafkaMessage: %s\n", kafkaMessage) //Message in single string
				// Start processing message
				ProcessKafkaMessage([]byte(kafkaMessage))
				if e.Headers != nil {
					log.Warnf("%% Headers: %v\n", e.Headers)
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
func ProcessKafkaMessage(rawMessage []byte) {
	// Kafka message convert to JSON struct
	message_slice := make([]gnfingest.PartDecodedMessage, 0)
	// Raw rawMessage can be an array of messages, or an object
	// Check for both or raise error if unknown
	var jsonErr error
	if len(rawMessage) > 0 && rawMessage[0] == '[' {
		// If the message is an array of object create slice of objects
		jsonErr = json.Unmarshal([]byte(rawMessage), &message_slice)
	} else if len(rawMessage) > 0 && rawMessage[0] == '{' {
		// If this is an object append to the slice of messages
		single_object := gnfingest.PartDecodedMessage{}
		jsonErr = json.Unmarshal([]byte(rawMessage), &single_object)
		message_slice = append(message_slice, single_object)
	} else { // Kafka message unknown format for unpacking
		jsonErr = errors.New("Raw kafka JSON message unknown format")
	}

	if jsonErr != nil { // Error Unable to decode JSON message
		log.Errorf("Kafka message Unmarshal error: %s. Msg: %s\n", jsonErr, rawMessage) //need to change raw message
	} else { // Iterate trhough slice of messages
		for _, part_decode_message := range message_slice {
			ProcessMessageHeader(part_decode_message)
		}
	}
}

func ProcessMessageHeader(partDecodeMessage gnfingest.PartDecodedMessage) {
	//convert partMessage into Message and rawTags Insert into message struct
	message := gnfingest.Message{}
	message.Name = partDecodeMessage.Name
	message.Timestamp = partDecodeMessage.Timestamp
	message.Values = partDecodeMessage.Values // Add values as rawBytes
	tagDecodeErr := json.Unmarshal(partDecodeMessage.Tags, &message.Tags)
	if tagDecodeErr != nil {
		log.Errorf("Message tags decode unmarshall errorr: %s Msg:%+v", tagDecodeErr, partDecodeMessage)
	}

	rawTags := partDecodeMessage.Tags // Extract the tags as raw bytes for decoding later

	msgVerify := message.MessageVerify() //Verify the message struct schema
	if msgVerify != nil {                // Not valid openconfig message
		log.Infof("JSON message invalid structure %s MSG: %s\n", msgVerify, (safeSlice((fmt.Sprintf("%s", message)), 400)))
	} else { // Valid openconfig message
		log.Infof("Valid JSON openconfig message: %s\n", (safeSlice((fmt.Sprintf("%s", message)), 400)))
		//Extract source IP and Path from message
		messageSource := message.MessageSource()
		messagePrefix := message.Tags.Prefix
		messagePath := message.Tags.Path
		messagePrefixPath := messagePrefix + messagePath
		log.Debugf("Message Source: %s Prefix: %s Path: %s\n", messageSource, messagePrefix, messagePath)
		//Start matching message to configured rules in config.json (device_keys)
		device, ok1 := Global_device_details[messageSource]
		if ok1 { // Source/Device IP is in config.json list
			log.Debugf("Device matched to message: %s\n", device.DeviceName)
			sensor, ok2 := device.Sensor[messagePrefixPath]
			if ok2 { // Match for Prefix and Path
				log.Debugf("Device & Sensor match to message: %s %+v\n", device.DeviceName, sensor)
				rule_id := sensor.KVS_rule_id //Rule-id for processing
				log.Infof("Processing rule Rule-ID: %s Name: %s SystemID: %s\n", rule_id, device.DeviceName, device.SystemID)
				ProcessJsonMessage(&message, rawTags, device, sensor)
			} else {
				log.Debugln("No matching sensor/rule in config.json")
			}
		} else {
			log.Debugln("No matching device in config.json")
		}
	}
}

func ProcessJsonMessage(message *gnfingest.Message, rawTags json.RawMessage, device *gnfingest.Device_Details, sensor gnfingest.Sensor) {
	//Rule-id for processing
	rule_id := sensor.KVS_rule_id
	rule := Rules[rule_id] // extract the rule in rules.json
	// Convert rawTags into mapped output for JSON searching
	var tagMap map[string]any
	err1 := json.Unmarshal(rawTags, &tagMap)
	if err1 != nil {
		log.Errorf("Message Tags to map Unmarshal error", err1)
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
	//log.Debugf("Index fields from message: %+v\n", fields)
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
}

func getFields(message *gnfingest.Message, fields map[string]interface{}, rule *gnfingest.YamlRule) {
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
