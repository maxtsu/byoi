package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"reflect"
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

	fmt.Printf(string(byteResult))
	//unmarshall to struct
	var msg Message
	err := json.Unmarshal(byteResult, &msg)
	if err != nil {
		panic(err)
	}
	//values in message
	prefix := msg.Prefix
	fmt.Printf("prefix %s\n", prefix)

	path := msg.Updates[0].Path
	fmt.Printf("path %s\n", path)

	raw := msg.Updates[0].Values

	var values map[string]interface{}
	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(raw), &values)
	// Print the data type
	fmt.Println(reflect.TypeOf(values))
	// Reading each value by its key
	fmt.Println("values :", values)

	gnfingest.Printme()

}
