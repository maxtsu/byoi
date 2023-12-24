package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gologme/log"
)

var batchSize = 10       // Influx write batch size
var flushInterval = 2000 // Influx write flush interval

var tand_host = "localhost"
var tand_port = "8086"
var database = "hb-default:cisco:cisco-B"
var measurement = "external/bt-kafka/cisco_resources/byoi"

// main function
func main() {
	// connect influxDB create Influx client return batchpoint
	var configfile = "config.json"
	//convert the config.json to a struct
	byteResult := gnfingest.ReadFile(configfile)
	var configjson gnfingest.Configjson
	err := json.Unmarshal(byteResult, &configjson)
	if err != nil {
		fmt.Println("config.json Unmarshall error", err)
	}
	// config.json list of device key from values under sensor for searching messages
	keys := []string{"path", "rule-id", "prefix"} //list of keys/parameters to extract from the KVS section
	device_details := configjson.DeviceDetails(keys)
	//Create InfluxDB client
	influxClient := gnfingest.InfluxdbClient(tand_host, tand_port, batchSize, flushInterval)
	log.Infof("Client create with client %+v\n", influxClient)
	fmt.Printf("Client: %+v\n", influxClient)
	gnfingest.InfluxClientWriteAPIs(influxClient, device_details)

	fmt.Printf("Device_details: %+v\n", device_details)

	// Create a point
	fields := map[string]interface{}{
		"source":       "nodeY",
		"admin-status": "UP",
		"oper-status":  "DOWN",
		"bytes_sent":   "234234",
	}

	//Write point to the writeAPI
	dev := device_details["10.213.94.44"]
	sensor := device_details["10.213.94.44"].Sensor["openconfig-interfaces:/interfaces/interface/state/"]
	fmt.Printf("Write points\n")

	times := time.Now()
	gnfingest.WritePoint(fields, times, &dev, &sensor)
	time.Sleep(3 * time.Second)
	fmt.Printf("Finish Write points\n")

	// Force all unwritten data to be sent
	//writeAPI.Flush()
	// Ensures background processes finishes
	//tandClient.Close()
}
