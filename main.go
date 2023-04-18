package main

import (
	"byoi/configjson"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var configfile = "/etc/byoi/config.json"

//var configfile = "config.json"

// Message struct
//{
//	"source": "node14:57400",
//	"subscription-name": "port_stats",
//	"timestamp": 1677602769296000000,
//	"time": "2023-02-28T16:46:09.296Z",
//	"prefix": "openconfig-interfaces:",
//	"updates": [

type Message struct {
	Source            string `json:"source"`
	Subscription_name string `json:"subscription-name"`
	Timestamp         int64  `json:"timestamp"`
	Time              string `json:"time"`
	Prefix            string `json:"prefix"`
	Updates           string `json:"updates"`
}

func main() {
	//convert the config.json to a struct
	configjson.ConfigJSON(configfile)

	// extract the brokers and topics from the configjson KVS
	brokertopic := configjson.KVS_parsing(configjson.Configuration.Hbin.Inputs[0].Plugin.Config.KVS, []string{"brokers", "topics"})
	fmt.Println("this is the broker: " + brokertopic[0])
	fmt.Println("this is the topics: " + brokertopic[1])

	devices := configjson.Configuration.Hbin.Inputs[0].Plugin.Config.Device
	// Call to get array of structs with device details
	list_devices := configjson.GetDevices(devices)
	fmt.Println("devices ")
	fmt.Println(list_devices[0].DeviceName)

	bootstrapServers := brokertopic[0]
	group := "byoi"
	topics := []string{brokertopic[1]}
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		// Avoid connecting to IPv6 brokers:
		// This is needed for the ErrAllBrokersDown show-case below
		// when using localhost brokers on OSX, since the OSX resolver
		// will return the IPv6 addresses first.
		// You typically don't need to specify this configuration property.
		//"broker.address.family": "v4",
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
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created Consumer %v\n", consumer)
	err = consumer.SubscribeTopics(topics, nil)

	run := true
	for run {
		fmt.Printf("waiting for kafka message\n")
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			// Poll the consumer for messages or events
			m := Message{}
			event := consumer.Poll(100)
			if event == nil {
				continue
			}
			switch e := event.(type) {
			case *kafka.Message:
				// Process the message received.
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				kafkaMessage := string(e.Value)
				sp = get_source_prefix(kafkaMessage)

				json.Unmarshal([]byte(kafkaMessage), &m)
				fmt.Printf("message struct: %+v\n", m)

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
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				// In this example we choose to terminate
				// the application if all brokers are down.
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}
}

type source_prefix struct {
	Source string
	Prefix string
}

func get_source_prefix(message string) source_prefix {
	m := source_prefix{}
	fmt.Println("Message: %s\n", message)
	json.Unmarshal([]byte(message), &m)
	println("Source: %s", m.Source)
	println("Prefix: %s", m.Prefix)
	return m
}
