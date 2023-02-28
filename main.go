package main

// this is a comment
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// var configfile = "/etc/byoi/config.json"
var configfile = "config.json"

// The Inputs struct for the config.json
type Hbin struct {
	Inputs  []Inputs  `json:"inputs"`
	Outputs []Outputs `json:"outputs"`
}

// The Outputs struct for the config.json
type Outputs struct {
	OutPlugin []OutPlugin `json:"plugin"`
}

// The Output Plugin struct for the config.json
type OutPlugin struct {
	Name      string      `json:"name"`
	OutConfig []OutConfig `json:"config"`
}

// The Config struct for the config.json
type OutConfig struct {
	Server   string `json:"server"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// The Inputs struct for the config.json
type Inputs struct {
	Plugin []Plugin `json:"plugin"`
}

// The Plugin struct for the config.json
type Plugin struct {
	Name   string   `json:"name"`
	Config []Config `json:"config"`
}

// The Config struct for the config.json
type Config struct {
	Device []Device `json:"device"`
	Devgrp string   `json:"device-group"`
	KVs    []KVs    `json:"kvs"`
}

// Device type
type Device struct {
	Name      string         `json:"name"`
	SystemID  string         `json:"system-id"`
	Sensor    []Sensor       `json:"sensor"`
	HBStorage HBStorage      `json:"healthbot-storage"`
	Auth      Authentication `json:"authentication"`
}

// Sensor config
type Sensor struct {
	Name        string `json:"name"`
	KVs         []KVs  `json:"kvs"`
	Measurement string `json:"measurement"`
}

// HBStorage to specify DB name and retention policy name for this device
type HBStorage struct {
	DB              string `json:"database"`
	RetentionPolicy string `json:"retention-policy"`
}

// Authentication config
type Authentication struct {
}

// Security config
type Security struct {
}

// KVs config
type KVs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func testConfig() {
	file, err := os.Open(configfile)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	defer file.Close() // it's important to close the file after reading it

	byteResult, _ := ioutil.ReadAll(file)

	var configuration Config
	json.Unmarshal(byteResult, &configuration)
	fmt.Println(configuration)
}

func configJSON() {
	// Let's first read the `config.json` file
	content, err := os.Open(configfile)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	defer fileContent.Close()
	byteResult, _ := ioutil.ReadAll(fileConten)
	fmt.Println(byteResult)
	// Now let's unmarshall the data into `byoiConfig`
	//	var Byoiconfig Hbin

	// Let's print the unmarshalled data!
	//log.Printf("Device: %s\n", ByoiConfig.Device)
	//	fmt.Println(Byoiconfig)
	// log.Printf("user: %s\n", byoiConfig.User)
	// log.Printf("status: %t\n", byoiConfig.Active)
}

func btkafka(broker string, topics []string, group string) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     broker,
		"broker.address.family": "v4",
		"group.id":              group,
		"session.timeout.ms":    6000,
		// Start reading from the first message of each assigned
		// partition if there are no previously committed offsets
		// for this group.x
		"auto.offset.reset": "earliest",
		// Whether or not we store offsets automatically.
		"enable.auto.offset.store": false,
	})

	if err != nil {
		fmt.Println("Failed to create consumer: %s\n", err)
	}
	fmt.Printf("Created Consumer %v\n", consumer)

	err = consumer.SubscribeTopics(topics, nil)

	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				// Process the message received.
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}
				_, err := consumer.StoreMessage(e)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%% Error storing offset after message %s:\n",
						e.TopicPartition)
				}
			case kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}
	fmt.Printf("Closing consumer\n")
	consumer.Close()
}

func main() {
	fmt.Println("Hello World!")
	//var broker = "10.54.162.129:9092"
	//var topics = []string{"cisco"}
	//var group = "healthbot"

	//read config file
	//configJSON()
	testConfig()

	//open connection to kafka broker
	// btkafka(broker, topics, group)
}
