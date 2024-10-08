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
)

// Global Variables
const batchSize = 600      // Influx write batch size
const flushInterval = 1000 // Influx write flush interval in ms
// const configfile = "/etc/byoi/config.json"
const configfile = "config.json"

var Global_device_details = make(map[string]*gnfingest.Device_Details)

// Getting Env details for TAND and group-id from ENV
var tand_host = (os.Getenv("TAND_HOST") + ".healthbot")
var tand_port = os.Getenv("TAND_PORT")
var group = (os.Getenv("CHANNEL") + "-golang1")

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

	// config.json list of parameters and values for each device
	device_details_slice := configjson.DeviceDetails()
	//Create Map of devices key is SystemID for source searching
	for _, d := range device_details_slice {
		Global_device_details[d.SystemID] = d //Add to global variable
	}

	//Create InfluxDB client
	gnfingest.InfluxCreateClient(tand_host, tand_port)
	log.Infof("Influx client create %+v\n", gnfingest.InfluxClient)

	// extract the brokers topics and configuration from the configjson KVS
	var kafkaCfg = make(map[string]string)
	kafkaCfg["brokers"] = gnfingest.KVS_parse_key(configjson.Hbin.Inputs[0].Plugin.Config.KVS, "brokers")
	kafkaCfg["topics"] = gnfingest.KVS_parse_key(configjson.Hbin.Inputs[0].Plugin.Config.KVS, "topics")
	kafkaCfg["saslusername"] = gnfingest.KVS_parse_key(configjson.Hbin.Inputs[0].Plugin.Config.KVS, "saslusername")
	kafkaCfg["saslpassword"] = gnfingest.KVS_parse_key(configjson.Hbin.Inputs[0].Plugin.Config.KVS, "saslpassword")
	kafkaCfg["saslmechanism"] = gnfingest.KVS_parse_key(configjson.Hbin.Inputs[0].Plugin.Config.KVS, "saslmechanism")
	kafkaCfg["securityprotocol"] = gnfingest.KVS_parse_key(configjson.Hbin.Inputs[0].Plugin.Config.KVS, "securityprotocol")

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

	//run := true
	run := false // for hometest
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
				log.Debugf("%% Message on %s: %s\n", e.TopicPartition, safeSlice(string(e.Value), 500))
				gnfingest.KafkaKeyCheck(e.Key) // search for device from the message key
				kafkaMessage := string(e.Value)
				// Start processing message in goroutine for concurrent processing
				ProcessKafkaMessage([]byte(kafkaMessage))
				if e.Headers != nil {
					log.Warnf("%% Headers: %v\n", e.Headers)
				}
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
	hometest() //call hometest function

	// Ensures background processes finishes
	gnfingest.InfluxClient.Close()
}

// Function to safely slice a string to a defined length
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
		log.Errorf("Kafka message Unmarshal error: %s. JSON Msg: %s\n", jsonErr, (safeSlice((fmt.Sprintf("%s", rawMessage)), 500)))
	} else { // Iterate trhough slice of messages
		for _, part_decode_message := range message_slice {
			//Basic check here to ensure event format message
			if len(part_decode_message.Tags) < 1 && len(part_decode_message.Values) < 1 {
				//This is not a valid event format message
				log.Infof("Not a valid event format JSON message: %s\n", (safeSlice((fmt.Sprintf("%s", rawMessage)), 500)))
			} else {
				ProcessMessageHeader(part_decode_message) //This is a valid Event format message
			}
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
		log.Errorf("Message Tags decode unmarshall error: %s Message:%+v", tagDecodeErr, partDecodeMessage)
	}
	rawTags := partDecodeMessage.Tags // Extract the tags as raw bytes for decoding later

	msgVerify := message.MessageVerify() //Verify the message struct schema
	if msgVerify != nil {                // Not valid openconfig message
		log.Infof("JSON message not valid telemetry event-format structure %s Message: %s\n", msgVerify, (safeSlice((fmt.Sprintf("%s", message)), 500)))
	} else { // Valid openconfig message
		log.Infof("Valid JSON openconfig telemetry event-format message: %s\n", (safeSlice((fmt.Sprintf("%s", message)), 500)))
		//Extract source IP and Path from message
		messageSource := message.MessageSource()
		messagePrefix := message.Tags.Prefix
		messagePath := message.Tags.Path
		messagePrefixPath := messagePrefix + messagePath
		log.Debugf("Message Source: %s Path: %s\n", messageSource, messagePrefixPath)

		//Start matching message to configured rules in config.json (device_keys)
		device, ok1 := Global_device_details[messageSource]
		if ok1 { // Source/Device IP is in config.json list
			log.Debugf("Device matched:%s Searching for sensor match\n", device.DeviceName)
			// Need to start a loop and loop through the sensors
			for _, sensor := range device.Sensor {
				//code to check if the sensor is a match
				if messagePrefixPath == sensor.PrefixPath {
					log.Debugf("Message matched Device:%s Matched Sensor:%s\n", device.DeviceName, sensor.Name)
					//code for sensor match
					ProcessJsonMessage(&message, rawTags, device, sensor)
				}
			}
		} else {
			log.Debugln("No matching device in config.json")
		}
	}
}

func ProcessJsonMessage(message *gnfingest.Message, rawTags json.RawMessage, device *gnfingest.Device_Details, sensor gnfingest.Sensor) {
	// Convert rawTags into mapped output for JSON searching
	var tagMap map[string]any
	err1 := json.Unmarshal(rawTags, &tagMap)
	if err1 != nil {
		log.Errorf("Message Tags to map Unmarshal error", err1)
	}
	// Declare fields map with interface
	var fields = make(map[string]interface{}) // fields & indexes from message
	// List of indexes from the matched sensor
	for _, i := range sensor.KVS_index {
		//Get indexes from Tags mapped value
		//Extract the index value from the Tags section of the message
		value := fmt.Sprintf("%v", tagMap[i]) //Convert all values into string
		fields[i] = value                     // Add value to data map
	}
	//Extract timestamp in Unix format to golang format
	t := time.Unix(0, message.Timestamp)
	//log.Debugf("Index fields from message: %+v\n", fields)
	// extract Field values
	getFields(message, fields, &sensor)
	// device_hostname added to fields for insertion into TSDB
	fields["device_hostname"] = device.Hostname
	log.Debugf("Data & Index fields from message: %+v\n", fields)

	// Create InfluxDB Point and append to slice/list in Device_details using the handler function
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

func getFields(message *gnfingest.Message, fields map[string]interface{}, sensor *gnfingest.Sensor) {
	// Receive raw data section of message (values) put in map
	var rawDataMap map[string]any
	err := json.Unmarshal(message.Values, &rawDataMap)
	if err != nil {
		log.Errorf("Kafka Values message Unmarshal error", err)
	}

	//Extract list of field values from the mapped values
	for _, fi := range sensor.KVS_fields {
		// The full field keyname is prefix and path
		fp := sensor.KVS_prefix + "/" + fi //add prefix to the field path
		//Extract the field value from the value section of the message
		value := fmt.Sprintf("%v", rawDataMap[fp]) //Convert all values into string
		fields[fi] = value                         //Add to fields map for point writting
	}
}

func hometest() {
	fmt.Printf("Home test\n")
	fmt.Printf("ALL DEVICES %+v\n", Global_device_details)
	for _, d := range Global_device_details { // Print all devices
		fmt.Printf("DEVICE: %+v\n", d)
		for _, s := range d.Sensor {
			fmt.Printf("SENSOR Name: %+v\n", s.Name)
			for _, f := range s.KVS_fields {
				fmt.Printf("fields: %+v\n", f)
			}
		}
	}

	// Processing sample message
	var sample = "ev-interface-state.json"
	//var sample = "interface-state.json"
	//var sample = "array.json"
	byteResult := gnfingest.ReadFile(sample)

	// fmt.Printf("message: %+v", kafkaMessage)
	fmt.Println("============Start ProcessKafkaMessage==========================")
	// processs message just like a kafka message
	fmt.Println(byteResult)
	go ProcessKafkaMessage(byteResult)
	// Wait for goroutine to finish
	time.Sleep(time.Second * 2)
}
