package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// var configfile = "/etc/byoi/config.json"
var configfile = "config.json"

func main() {
	//convert the config.json to a struct
	byteResult := gnfingest.ReadFile(configfile)
	var configuration gnfingest.JSONfile
	err := json.Unmarshal(byteResult, &configuration)
	if err != nil {
		fmt.Println("Unmarshall error", err)
	}
	// extract the brokers and topics from the configjson KVS
	brokertopic := gnfingest.KVS_parsing(configuration.Hbin.Inputs[0].Plugin.Config.KVS, []string{"brokers", "topics"})

	fmt.Println("this is the broker: " + brokertopic[0])
	fmt.Println("this is the topics: " + brokertopic[1])

	//list of devices configuration from configjson
	bootstrapServers := brokertopic[0]
	group := "byoi"
	topics := []string{brokertopic[1]}
	devices := configuration.Hbin.Inputs[0].Plugin.Config.Device
	//list of device key values under sensor for searching messages
	keys := []string{"prefix", "path"}
	device_keys := gnfingest.DeviceDetails(devices, keys)

	fmt.Printf("Device-Keys %+v\n", device_keys)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
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

	//run := true
	run := false
	for run {
		fmt.Printf("waiting for kafka message\n")
		time.Sleep(2 * time.Second)
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			// Poll the consumer for messages or events
			m := gnfingest.Message{}
			event := consumer.Poll(400)
			if event == nil {
				continue
			}
			switch e := event.(type) {
			case *kafka.Message:
				// Process the message received.
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				kafkaMessage := string(e.Value)

				//sp := get_source_prefix(kafkaMessage)

				json.Unmarshal([]byte(kafkaMessage), &m)
				fmt.Printf("message struct: %+v\n", m)

				//Start matching message to configured rules
				//m := message_root{}
				for _, d := range device_keys {
					fmt.Println("Device: ", d.DeviceName)
					if (d.DeviceName == m.Source) && (d.KVS_prefix == m.Prefix) {
						fmt.Printf("name and prefix match ")
					}
				}

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

	// Processing sample message
	//var sample = "Sample4.json"
	//var sample = "Sample20.json"
	var sample = "interface-state.json"
	//var sample = "isis-1.json"
	byteResult = gnfingest.ReadFile(sample)
	var kafkaMessage gnfingest.Message
	err = json.Unmarshal(byteResult, &kafkaMessage)
	if err != nil {
		fmt.Println("Unmarshall error", err)
	}

	// extract source and prefix
	fmt.Println("source: %s", kafkaMessage.Source)
	fmt.Println("prefix: %s", kafkaMessage.Prefix)
	fmt.Println("path: %s", kafkaMessage.Updates[0].Path)

	// parse to extract path & indexes from "Path:" value in message
	var result []string
	var k1 = make(map[string][]gnfingest.KVS)
	result, k1 = gnfingest.PathExtract(kafkaMessage.Updates[0].Path)
	fmt.Println("path list: %+v\n", result)
	fmt.Println("map of keys: %+v\n", k1)

	// Test what struct branch is created
	var ZeroValues gnfingest.Values
	fmt.Printf("Message: %+v\n", kafkaMessage)
	fmt.Printf("path: %+v\n", kafkaMessage.Updates[0].Values.State)

	if kafkaMessage.Updates[0].Values.Counters != ZeroValues.Counters {
		fmt.Println("This is counters message\n")
	}
	if kafkaMessage.Updates[0].Values.Isis.Interfaces.Interface.InterfaceID != ZeroValues.Isis.Interfaces.Interface.InterfaceID {
		fmt.Println("This is isis message\n")
		fmt.Printf("Message: %+v\n", kafkaMessage.Updates[0].Values.Isis)
	}
	if kafkaMessage.Updates[0].Values.State != ZeroValues.State {
		fmt.Println("This is interface-state message\n")
	}
}
