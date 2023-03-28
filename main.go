package main

// Program calls the 3 components for the healthbot plugin
// Kafka module for ingesting data
// configjson for reading the /etc/byoi/config.json
// influxdb point write

import (
	"fmt"
	"maxwell/configjson"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// the plugin config.json file
// var configfile = "/etc/byoi/config.json"
var configfile = "config.json"

func main() {
	fmt.Println("HelloWorld!")
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
	fmt.Println(list_devices[0].SensorPath)

	bootstrapServers := brokertopic[0]
	group := "byoi"
	topics := brokertopic[1]
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		// Avoid connecting to IPv6 brokers:
		// This is needed for the ErrAllBrokersDown show-case below
		// when using localhost brokers on OSX, since the OSX resolver
		// will return the IPv6 addresses first.
		// You typically don't need to specify this configuration property.
		"broker.address.family": "v4",
		"group.id":              group,
		"session.timeout.ms":    6000,
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
	//run := true
}
