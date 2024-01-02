package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gologme/log"
)

// Global Variables
const BatchSize = 10       // Influx write batch size
const flushInterval = 2000 // Influx write flush intervale

var tand_host = "localhost"
var tand_port = "8086"

// var tand_host = (os.Getenv("TAND_HOST") + ".healthbot")
// var tand_port = os.Getenv("TAND_PORT")
var database = "hb-default:cisco:cisco-B"
var measurement = "external/bt-kafka/cisco_resources/byoi"

// main function
func main() {
	// connect influxDB create Influx client return batchpoint
	var configfile = "config.json"
	//var configfile = "/etc/byoi/config.json"
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
	InfluxClient := gnfingest.InfluxCreateClient(tand_host, tand_port)
	log.Infof("Client create with client %+v\n", InfluxClient)
	fmt.Printf("Client: %+v\n", InfluxClient)

	fmt.Printf("Device_details: %+v\n", device_details)
	//Write point to the writeAPI
	dev := device_details["10.213.94.44"]
	sensor := device_details["10.213.94.44"].Sensor["openconfig-interfaces:/interfaces/interface/state/"]

	// Create a point
	fields := map[string]interface{}{
		"interfaces/interface/name":                "ge-0/0/1",
		"interfaces/interface/status/admin-status": "UP",
		"oper-status": "DOWN",
	}
	times := time.Now()
	dev.AddPoint(fields, times, &sensor, batchSize)

	// Create a 2nd point
	fields = map[string]interface{}{
		"interfaces/interface/name":                "ge-0/0/2",
		"interfaces/interface/status/admin-status": "UP",
	}
	times = time.Now()
	dev.AddPoint(fields, times, &sensor, batchSize)
	//gnfingest.AddPoint(fields, times, &dev, &sensor)

	pts := dev.Points
	fmt.Printf("Listed Points %+v\n", pts)

	// Flush points
	fmt.Printf("Device Points before %+v\n", dev.Points)
	dev.FlushPoints(InfluxClient)
	fmt.Printf("Device Points after %+v\n", dev.Points)

	time.Sleep(5 * time.Second)
	fmt.Printf("Finish Write points\n")

}
