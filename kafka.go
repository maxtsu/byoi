package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const config_file = "kafka-conf.json"

func main() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	byteResult := gnfingest.ReadFile(config_file)
	var configjson Config
	err := json.Unmarshal(byteResult, &configjson)
	if err != nil {
		fmt.Println("app-config.json Unmarshall error", err)
	}
	fmt.Printf("app-config.json %+v\n", configjson)
	// Create kafka consumer configuration for kafkaCfg
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  configjson.Kafka.BootstrapServers,
		"sasl.mechanisms":    configjson.Kafka.SaslMechanisms,
		"security.protocol":  configjson.Kafka.SecurityProtocol,
		"sasl.username":      configjson.Kafka.SaslUsername,
		"sasl.password":      configjson.Kafka.SaslPassword,
		"ssl.ca.location":    configjson.Kafka.SslCaLocation,
		"group.id":           configjson.Kafka.GroupID,
		"session.timeout.ms": 6000,
		// Start reading from the first message of each assigned
		// partition if there are no previously committed offsets
		// for this group.
		"auto.offset.reset": configjson.Kafka.AutoOffset,
		// Whether or not we store offsets automatically.
		"enable.auto.offset.store": false,
	})
	if err != nil {
		fmt.Println("Failed to create consumer. ", err)
		os.Exit(1)
	}
	fmt.Println("Created Consumer. ", consumer)

	topics := []string{configjson.Kafka.Topics}
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
			message := gnfingest.Message{}
			event := consumer.Poll(400)
			if event == nil {
				continue
			}
			switch e := event.(type) {
			case *kafka.Message:
				// Process the message received.
				//fmt.Printf("Got a kafka message\n")
				fmt.Printf("%% Message on %s: %s\n", e.TopicPartition, string(e.Value)[100:])
				kafkaMessage := string(e.Value)
				fmt.Printf("\nkafkaMessage: %s\n", kafkaMessage) //Message in single string
				json.Unmarshal([]byte(kafkaMessage), &message)
				// Start processing message
				//ProcessKafkaMessage(&message, device_keys)
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
				fmt.Printf("%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					fmt.Printf("Kafka error. All brokers down ")
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}
}

// configuration file app-config.json
type Config struct {
	Kafka struct {
		BootstrapServers string `json:"bootstrap.servers"`
		SaslMechanisms   string `json:"sasl.mechanisms"`
		SecurityProtocol string `json:"security.protocol"`
		SaslUsername     string `json:"sasl.username"`
		SaslPassword     string `json:"sasl.password"`
		SslCaLocation    string `json:"ssl.ca.location"`
		GroupID          string `json:"group.id"`
		Topics           string `json:"topics"`
		AutoOffset       string `json:"auto.offset.reset"`
	} `json:"kafka"`
}
