package main

import (
    "context"
    "fmt"
    "os"
    "time"

    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)


func main() {
	if err := dbWrite(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

}

func dbWrite(ctx context.Context) error {
    // Create write client
    url := "https://us-east-1-1.aws.cloud2.influxdata.com"
    token := os.Getenv("INFLUXDB_TOKEN")
    writeClient := influxdb2.NewClient(url, token)

    // Define write API
    org := "Dev"
    bucket := "test"
    writeAPI := writeClient.WriteAPIBlocking(org, bucket)

data := map[string]map[string]interface{}{
  "point1": {
    "location": "Klamath",
    "species":  "bees",
    "count":    23,
  },
  "point2": {
    "location": "Portland",
    "species":  "ants",
    "count":    30,
  },
  "point3": {
    "location": "Klamath",
    "species":  "bees",
    "count":    28,
  },
  "point4": {
    "location": "Portland",
    "species":  "ants",
    "count":    32,
  },
  "point5": {
    "location": "Klamath",
    "species":  "bees",
    "count":    29,
  },
  "point6": {
    "location": "Portland",
    "species":  "ants",
    "count":    40,
  },
}

// Write data
for key := range data {
  point := influxdb2.NewPointWithMeasurement("census").
    AddTag("location", data[key]["location"].(string)).
    AddField(data[key]["species"].(string), data[key]["count"])

  if err := writeAPI.WritePoint(ctx, point); err != nil {
    return fmt.Errorf("write API write point: %s", err)
  }

  time.Sleep(1 * time.Second) // separate points by 1 second
}

return nil
}
