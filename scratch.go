package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

// main function
func main() {
	var sample = "interface-state.json"
	byteResult := gnfingest.ReadFile(sample)
	var kafkaMessage gnfingest.Message
	err := json.Unmarshal(byteResult, &kafkaMessage)
	if err != nil {
		fmt.Println("Unmarshall error", err)
	}
	// extract source and prefix
	fmt.Printf("Message: %+v\n", kafkaMessage)
	fmt.Println("source: %s", kafkaMessage.Source)
	fmt.Println("timestamp: %s", kafkaMessage.Timestamp)
	fmt.Println("prefix: %s", kafkaMessage.Prefix)
	fmt.Println("path: %s", kafkaMessage.Updates[0].Path)

	//var source = kafkaMessage.Source
	//var timestamp = kafkaMessage.Timestamp

	fmt.Printf("\ntimestamp: %+v\n", kafkaMessage.Timestamp)

	//unixTime := time.Unix(0, timestamp) //gives unix time stamp in utc

	//ExampleClient_query(source, unixTime)

	// Load rules.json into struct
	var rulesfile = "rules.json"
	byteResult = gnfingest.ReadFile(rulesfile)
	var r []gnfingest.RulesJSON
	err = json.Unmarshal(byteResult, &r)
	if err != nil {
		fmt.Println("Unmarshall error", err)
	}
	//fmt.Printf("Rules:  %+v\n", rules)
	// create map of structs key=rule-id
	var rules = make(map[string]gnfingest.RulesJSON)
	for _, r := range r {
		rules[r.RuleID] = r
	}
	fmt.Printf("Rules:  %+v\n", rules)
}

// start a new client
func ExampleClient_query(source string, timestamp time.Time) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://10.54.182.2:8086",
	})
	fmt.Printf("opened client\n")
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database: "TEST01",
		//Precision: "s",
	})

	// Create a point and add to batch
	tags := map[string]string{"source": source}
	fields := map[string]interface{}{
		"admin-status": "UP",
		"oper-status":  "DOWN",
	}
	pt, err := client.NewPoint("cpu_usage", tags, fields, time.Now())
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	bp.AddPoint(pt)

	// Write the batch
	c.Write(bp)

}
