package main

import (
	"fmt"
	"time"

	"github.com/gologme/log"
	client "github.com/influxdata/influxdb/client/v2"
)

// main function
func main() {
	// connect influxDB create Influx client return batchpoint
	//tand_host := "localhost"
	//tand_port := "8086"

	InfluxDB2()

	// Write the batch test vv
	/*err = tandClient.Write(batchPoint)
	if err != nil {
		fmt.Println("Write Error: ", err.Error())
	} else {
		fmt.Println("Succesful write")
	}*/
}

func InfluxDB2() {
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created InfluxDB Client: %+v\n", c)
	defer c.Close()

	// Create a new point batch
	MyDB := "hb-default:cisco:cisco-B"
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MyDB,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
		fmt.Printf("err: \n")
	}
	fmt.Printf("Created batchP: %+v\n", bp)
	// Create a point and add to batch
	tags := map[string]string{"cpu": "cpu-total"}
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	}

	pt, err := client.NewPoint("cpu_usage", tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
		fmt.Printf("err: \n")
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Fatal(err)
		fmt.Printf("err: \n")
	} else {
		fmt.Printf("Succesful write: \n")
	}

	// Close client resources
	if err := c.Close(); err != nil {
		log.Fatal(err)
		fmt.Printf("err: \n")
	}
}
