package main

import (
	"byoi/gnfingest"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gologme/log"
	client "github.com/influxdata/influxdb1-client/v2"
)

//var batchSize = 10       // Influx write batch size
//var flushInterval = 2000 // Influx write flush interval

// var tand_host = "localhost"
// var tand_port = "8086"
var tand_host = (os.Getenv("TAND_HOST") + ".healthbot")
var tand_port = os.Getenv("TAND_PORT")
var database = "hb-default:cisco:cisco-B"
var measurement = "external/bt-kafka/cisco_resources/byoi"

// main function
func main() {
	// connect influxDB create Influx client return batchpoint
	var configfile = "config.jsonX"
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
	influxClient := gnfingest.InfluxdbClient(tand_host, tand_port)
	log.Infof("Client create with client %+v\n", influxClient)
	fmt.Printf("Client: %+v\n", influxClient)
	// Create Batch Points
	gnfingest.InfluxClientBatchPoint(influxClient, device_details)

	fmt.Printf("Device_details: %+v\n", device_details)
	//Write point to the writeAPI
	dev := device_details["192.168.1.22"]
	sensor := device_details["192.168.1.22"].Sensor["openconfig-interfaces:/interfaces/interface/state/"]
	fmt.Printf("Write points\n")

	fmt.Printf("Empty BatchP %+v\n", dev.BatchPoint)
	// Create a point
	fields := map[string]interface{}{
		"interfaces/interface/name":                "ge-0/0/1",
		"interfaces/interface/status/admin-status": "UP",
		"oper-status": "DOWN",
	}
	times := time.Now()
	gnfingest.AddPoint(fields, times, &dev, &sensor)

	// Create a 2nd point
	fields = map[string]interface{}{
		"interfaces/interface/name":                "ge-0/0/2",
		"interfaces/interface/status/admin-status": "UP",
	}
	times = time.Now()
	gnfingest.AddPoint(fields, times, &dev, &sensor)

	pts := dev.BatchPoint.Points()
	fmt.Printf("Listed Points %+v\n", pts)

	// Write the batch
	fmt.Printf("BatchPoints before %+v\n", dev.BatchPoint)
	WriteBatch(influxClient, dev.BatchPoint)
	influxClient.Write(dev.BatchPoint)

	fmt.Printf("BatchPoints after %+v\n", dev.BatchPoint)
	time.Sleep(8 * time.Second)
	fmt.Printf("Finish Write points\n")

}

func WriteBatch(c client.Client, bp client.BatchPoints) {
	c.Write(bp)
}
