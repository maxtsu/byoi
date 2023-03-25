package main

// Test program for linux compile
import (
	"configjson"
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	fmt.Println("HelloWorld!")
	configjson.ConfigJSON("config.json")

	fmt.Println("hbDB: " + configjson.Configuration.Hbin.Inputs[0].Plugin.Config.Device[0].HBStorage.DB)

	var broker = "10.54.162.129:9092"
	var topics = []string{"cisco"}
	var group = "healthbot"
	var run = true

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     broker,
		"broker.address.family": "v4",
		"group.id":              group,
		"session.timeout.ms":    6000,
		// Start reading from the first message of each assigned
		// partition if there are no previously committed offsets
		// for this group.x
		"auto.offset.reset": "earliest",
		// Whether or not we store offsets automatically
		"enable.auto.offset.store": false,
	})

	err = consumer.SubscribeTopics(topics, nil)

	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	for run == true {
		ev := consumer.Poll(100)
		switch e := ev.(type) {
		case *kafka.Message:
			// application-specific processing
		case kafka.Error:
			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			run = false
		default:
			fmt.Printf("Ignored %v\n", e)
		}
	}

	consumer.Close()

}
