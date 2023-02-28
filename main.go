package main

// this is a comment
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// var configfile = "/etc/byoi/config.json"
var configfile = "config.json"

//var configfile = "users.json"

// The Top struct for the config.json
type Top struct {
	Hbin    Hbin    `json:"hbin"`
	Logging Logging `json:"logging"`
}

// The Logging struct for the config.json
type Logging struct {
}

// The Hbin struct for the config.json
type Hbin struct {
	Inputs  Inputs  `json:"inputs"`
	Outputs Outputs `json:"outputs"`
}

// The Outputs struct for the config.json
type Outputs struct {
	// OutPlugin []OutPlugin `json:"plugin"`
}

// The Inputs struct for the config.json
type Inputs struct {
	// Plugin []Plugin `json:"plugin"`
}

// Users struct which contains
// an array of users
type Users struct {
	Users []User `json:"users"`
}

// User struct which contains a name
// a type and a list of social links
type User struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Age    int      `json:"Age"`
	Social []Social `json:"social"`
}

// Social struct which contains a
// list of links
type Social struct {
	// Facebook string `json:"facebook"`
	// Twitter  string `json:"twitter"`
}

func configJSON() {
	file, err := os.Open(configfile)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	defer file.Close()

	byteResult, _ := ioutil.ReadAll(file)

	//var res map[string]interface{}
	//json.Unmarshal([]byte(byteResult), &res)
	//fmt.Println(res)

	// we initialize our Users array
	var users Users
	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteResult, &users)
	fmt.Println(users)
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
	configJSON()

	//open connection to kafka broker
	// btkafka(broker, topics, group)
}
