package main

import (
	"fmt"
	"time"

	"github.com/gologme/log"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// main function
func main() {
	// connect influxDB create Influx client return batchpoint
	tand_host := "localhost"
	tand_port := "8086"
	database := "hb-default:cisco:cisco-B"
	measurement := "external/bt-kafka/cisco_resources/byoi"
	//Create client
	tandClient := InfluxdbClient(tand_host, tand_port)
	fmt.Printf("Client create with client %+v\n", tandClient)

	//writeAPI := WriteApi(database, tandClient)
	writeAPI := tandClient.WriteAPI("my-org", database)
	// Create a point
	tags := map[string]string{}
	fields := map[string]interface{}{
		"source":       "nodeX",
		"admin-status": "UP",
		"oper-status":  "DOWN",
	}

	p := influxdb2.NewPoint(
		measurement, tags, fields, time.Now(),
	)
	/*
	   p := influxdb2.NewPoint("stat",
	       map[string]string{"unit": "temperature"},
	       map[string]interface{}{"avg": 24.5, "max": 45.0},
	       time.Now()) */

	writeAPI.WritePoint(p)
	// Force all unwritten data to be sent
	writeAPI.Flush()
	// Ensures background processes finishes
	tandClient.Close()
}

func InfluxdbClient(tand_host string, tand_port string) influxdb2.Client {
	url := "http://" + tand_host + ":" + tand_port
	//config := client.HTTPConfig{Addr: url}
	//c, err := client.NewHTTPClient(config)
	// create client
	/*	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: url,
	}) */

	c := influxdb2.NewClientWithOptions(url, "my-token",
		influxdb2.DefaultOptions().SetBatchSize(20))
	defer c.Close()
	log.Infof("Created InfluxDB Client: %+v\n", c)
	return c //return the client
}
