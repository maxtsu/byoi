package main

import (
	"fmt"
	"time"

	"github.com/gologme/log"
	client "github.com/influxdata/influxdb1-client/v2"
)

// main function
func main() {
	// connect influxDB create Influx client return batchpoint
	tand_host := "localhost"
	tand_port := "8086"
	database := "hb-default:cisco:cisco-B"
	measurement := "external/bt-kafka/cisco_resources/byoi"
	tandClient := InfluxdbClient(tand_host, tand_port)
	fmt.Printf("Client create with client %+v\n", tandClient)

	//create batch point with database
	batchPoint := DatabaseBp(database)
	fmt.Printf("Client create with BP %+v\n", batchPoint)

	// Create a point and add to batch
	tags := map[string]string{}
	fields := map[string]interface{}{
		"source":       "nodeX",
		"admin-status": "UP",
		"oper-status":  "DOWN",
	}
	pt, err := client.NewPoint(measurement, tags, fields, time.Now())
	if err != nil {
		fmt.Println("New point Error: ", err.Error())
	} else {
		fmt.Println("Created point: ", pt)
	}
	batchPoint.AddPoint(pt)

	// Write the batch test
	err = tandClient.Write(batchPoint)
	if err != nil {
		fmt.Println("Write Error: ", err.Error())
	} else {
		fmt.Println("Succesful write")
	}
}

// create InfluxDB v1.8 client
func InfluxdbClient(tand_host string, tand_port string) client.Client {
	url := "http://" + tand_host + ":" + tand_port
	config := client.HTTPConfig{Addr: url}
	// create client
	/*c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: url,
	})*/
	c, err := client.NewHTTPClient(config)
	if err != nil {
		log.Errorf("Error creating InfluxDB Client: ", err)
	}
	defer c.Close()
	log.Infof("Created InfluxDB Client: %+v\n", c)
	return c //return the client
}

// create InfluxDB batchpoint with database
func DatabaseBp(database string) client.BatchPoints {
	// Create a new point batch for database
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database: database,
		//Precision: "s",
	})
	//fmt.Printf("Client create with BP %+v %+v", bp, c)
	return bp
}
