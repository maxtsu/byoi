package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"

	client "github.com/influxdata/influxdb1-client/v2"
)

// declaring a struct
type Message struct {
	Source           string    `json:"source"`
	SubscriptionName string    `json:"subscription-name"`
	Timestamp        int64     `json:"timestamp"`
	Time             string    `json:"time"`
	Prefix           string    `json:"prefix"`
	Updates          []Updates `json:"updates"`
}

type Updates struct {
	Path   string          `json:"path"`
	Values json.RawMessage `json:"values"`
}

// main function
func main() {
	// json file
	var sample = "Sample4.json"

	byteResult := gnfingest.ReadFile(sample)

	//fmt.Printf(string(byteResult))
	//unmarshall to struct
	var msg Message
	err := json.Unmarshal(byteResult, &msg)
	if err != nil {
		panic(err)
	}
	//values in message
	prefix := msg.Prefix
	fmt.Printf("prefix %s\n", prefix)
	ExampleClient_query()
}

// start a new client
func ExampleClient_query() {
	fmt.Printf("opening client\n")
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://10.54.182.2:8086",
	})
	fmt.Printf("opened client\n")
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	q := client.NewQuery("SELECT * FROM external/bt-kafka/cisco_resources/byoi", "TEST01", "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		fmt.Println(response.Results)
	} else {
		fmt.Printf("not in DB\n")
	}
}
