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
var flushInterval = 2000 // Influx write flush intervale

// main function
func main() {
	// connect influxDB create Influx client return batchpoint
	tand_host := "localhost"
	tand_port := "8086"
	database := "hb-default:cisco:cisco-B"
	measurement := "external/bt-kafka/cisco_resources/byoi"

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

	//Create client
	tandClient := InfluxdbClientx(tand_host, tand_port)
	fmt.Printf("Client create with client %+v\n", tandClient)

	//Create InfluxDB client
	influxClient := InfluxdbClientx(tand_host, tand_port)
	log.Infof("Client create with client %+v\n", influxClient)
	fmt.Printf("Client: %+v\n", influxClient)

	// Create Influx writeAPI for each database (source) from device_details list
	for _, d := range device_details {
		databas := d.Database
		wapi := influxClient.WriteAPI("my-org", databas)
		fmt.Printf("wapi: %+v\n", wapi)
		d.WriteApi = wapi
		d.Test = d.DeviceName
		fmt.Printf("d.WriteApi: %+v\n", d.WriteApi)
		fmt.Printf("d.Test: %+v\n", d.Test)
	}
	fmt.Printf("\nPrinting the wrtieapi again\n")
	for _, d := range device_details {
		fmt.Printf("d.WriteApi: %+v\n", d.WriteApi)
		fmt.Printf("d.Test: %+v\n", d.Test)
	}

	//writeAPI := WriteApi(database, tandClient)
	writeAPI := tandClient.WriteAPI("my-org", database)

	fmt.Printf("Type of %v is %T", writeAPI, writeAPI)

	for i := 1; i < 3; i++ {
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
		writeAPI.WritePoint(p)
		fmt.Printf("Write points: %+v\n", i)
		time.Sleep(3 * time.Second)
	}
	fmt.Printf("Write points flush\n")
	// Force all unwritten data to be sent
	writeAPI.Flush()
	// Ensures background processes finishes
	tandClient.Close()
}

// create InfluxDB client
func InfluxdbClientx(tand_host string, tand_port string) influxdb2.Client {
	// set options for influx client
	options := influxdb2.DefaultOptions()
	options.SetBatchSize(uint(batchSize))
	options.SetFlushInterval(uint(flushInterval))
	options.SetLogLevel(2) //0 error, 1 - warning, 2 - info, 3 - debug

	// create client
	url := "http://" + tand_host + ":" + tand_port
	c := influxdb2.NewClientWithOptions(url, "my-token", options)
	defer c.Close()
	log.Infof("Created InfluxDB Client: %+v\n", c)
	return c //return the influx client
}

// Create the point with data for writing
func WritePointx(fields map[string]interface{}, msg *gnfingest.Message, dev *gnfingest.Device_Details) {
	tags := map[string]string{}
	time := time.Unix(msg.Timestamp, 0)
	p := influxdb2.NewPoint(dev.Measurement, tags, fields, time)
	fmt.Printf("Point: %+v\n", p)
	if dev.WriteApi != nil {
		//Write point to the writeAPI
		dev.WriteApi.WritePoint(p)
		log.Debugf("Write data point: %+v\n", p)
	} else {
		log.Errorf("WriteApi for: %+v <nil>\n", dev.DeviceName)
	}
}
