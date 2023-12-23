package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gologme/log"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
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
	tags := map[string]string{}
	fields := map[string]interface{}{
		"source":       "nodeX",
		"admin-status": "UP",
		"oper-status":  "DOWN",
		"bytes_sent":   i,
	}

	p := influxdb2.NewPoint(
		measurement, tags, fields, time.Now(),
	)
	//Write point to the writeAPI
	dev := device_details["10.213.94.44"]
	sensor := device_details["10.213.94.44"].Sensor["openconfig-interfaces:/interfaces/interface/state/"]
	fmt.Printf("Write points\n")
	//time.Sleep(3 * time.Second)

	time := time.Now()
	gnfingest.WritePoint(fields, time, &dev, &sensor)

	// Force all unwritten data to be sent
	//writeAPI.Flush()
	// Ensures background processes finishes
	//tandClient.Close()
}
