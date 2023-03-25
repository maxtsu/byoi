package main

// Test program for linux compile
import (
	"fmt"
	"maxwell/configjson"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	fmt.Println("HelloWorld!")
	configjson.ConfigJSON("config.json")

	fmt.Println("hbDB: " + configjson.Configuration.Hbin.Inputs[0].Plugin.Config.Device[0].HBStorage.DB)

	var broker = "10.54.162.129:9092"
	var topics = []string{"cisco"}
	var group = "healthbot"

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

}
