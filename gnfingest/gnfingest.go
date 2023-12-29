package gnfingest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gologme/log"
	client "github.com/influxdata/influxdb1-client/v2"
)

var InfluxClient client.Client

// gnmic Event Message partial struct
type Message struct {
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
	Tags      struct {
		Path             string `json:"path"`
		Prefix           string `json:"prefix"`
		Source           string `json:"source"`
		SubscriptionName string `json:"subscription-name"`
	} `json:"tags"`
	Values json.RawMessage `json:"values"`
}

// Message Method to verify openconfig JSON event message
func (m *Message) MessageVerify() error {
	if m.Name == "" {
		return fmt.Errorf("no Name field in message")
	} else if m.Tags.Prefix == "" {
		return fmt.Errorf("no Prefix field in message")
	} else if m.Tags.Source == "" {
		return fmt.Errorf("no Source field in message")
	} else if string(m.Tags.Path) == "" {
		return fmt.Errorf("no Path field in message")
	} else if string(m.Values) == "" {
		return fmt.Errorf("no Values field in message")
	} else { // This openconfig message is OK
		return nil
	}
}

// Message Method to extract source IP & path
func (m *Message) MessageSource() string {
	// Extract message source IP remove port number
	var source string
	if strings.Contains(m.Tags.Source, ":") {
		s := strings.Split(m.Tags.Source, ":")
		source = s[0]
	} else {
		source = m.Tags.Source
	}
	return source
}

// gnmic Event Message Tags only as raw data
type MessageTags struct {
	Tags json.RawMessage `json:"tags"`
}

// The Top struct for the config.json
type Configjson struct {
	Hbin    Hbin `json:"hbin"`
	Logging struct {
		Level   string `json:"level"`
		Enabled string `json:"enabled"`
	} `json:"logging"`
}

// The Hbin struct for the config.json
type Hbin struct {
	Inputs  []Inputs  `json:"inputs"`
	Outputs []Outputs `json:"outputs"`
}

type Outputs struct {
	Plugin struct {
		Name   string `json:"name"`
		Config struct {
			Server   string `json:"server"`
			Port     int    `json:"port"`
			User     string `json:"user"`
			Password string `json:"password"`
		} `json:"config"`
	} `json:"plugin"`
}

// The Inputs struct for the config.json
type Inputs struct {
	Plugin struct {
		Name   string `json:"name"`
		Config struct {
			Device      []Device `json:"device"`
			Devicegroup string   `json:"device-group"`
			KVS         []KVS    `json:"kvs"`
		} `json:"config"`
	} `json:"plugin"`
}

// Device type
type Device struct {
	Name   string `json:"name"`
	Sensor []struct {
		Name        string `json:"name"`
		KVS         []KVS  `json:"kvs"`
		Measurement string `json:"measurement"`
	} `json:"sensor"`
	Authentication struct {
		Password struct {
			Password string `json:"password"`
			Username string `json:"username"`
		} `json:"password"`
	} `json:"authentication"`
	HealthbotStorage struct {
		Database        string `json:"database"`
		RetentionPolicy string `json:"retention-policy"`
	} `json:"healthbot-storage"`
	SystemID string `json:"system-id"`
}

// KVS struct
type KVS struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Function to return list/slice of device details from config.json
func (c *Configjson) DeviceDetails(keys []string) map[string]Device_Details {
	// create map of devices key is DeviceName
	var device_details = make(map[string]Device_Details)
	for _, d := range c.Hbin.Inputs[0].Plugin.Config.Device {
		var dev Device_Details
		dev.DeviceName = d.Name
		dev.SystemID = d.SystemID
		dev.Database = d.HealthbotStorage.Database
		var sensors = map[string]Sensor{}
		//extract list/slice of structs for sensors
		//d.MapKVS() //Create map for KVS items
		// Iterate over array of sensors
		for _, s := range d.Sensor {
			var sensor Sensor
			sensor.Measurement = s.Measurement
			kvs_pairs := KVS_parsing(s.KVS, keys)
			// Parameters from config.json {path}
			sensor.KVS_path = kvs_pairs["path"]
			sensor.KVS_rule_id = kvs_pairs["rule-id"]
			sensor.KVS_prefix = kvs_pairs["prefix"]
			//Sensor map index is concatenated prefix + path
			map_key := sensor.KVS_prefix + sensor.KVS_path
			sensors[map_key] = sensor
		}
		dev.Sensor = sensors
		//Device map key is SystemID for source searching
		device_details[dev.SystemID] = dev
	}
	return device_details
}

// struct defining sensor/rule for each device
//
// -device name -kvs path -kvs rule-id
// -kvs prefix -measurement -database
type Device_Details struct {
	DeviceName string
	Database   string
	SystemID   string
	Points     []*client.Point
	Sensor     map[string]Sensor //key for map is KVS_path
}

// struct defining sensor/rule for each device
type Sensor struct {
	KVS_path    string
	KVS_rule_id string
	KVS_prefix  string
	Measurement string
}

// Function to read text file return byteResult
func ReadFile(fileName string) []byte {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("File reading error", err)
		return []byte{}
	}
	byteResult, _ := io.ReadAll(file)
	file.Close()
	return byteResult
}

// Function to extract KVS pairs by key names passed into function
func KVS_parsing(keys []KVS, keyString []string) map[string]string {
	// Build a config map:
	confMap := map[string]string{}
	for _, v := range keys {
		confMap[v.Key] = v.Value
	}
	// Find values by key in the config map
	var results = make(map[string]string) //map for return values
	for _, key := range keyString {
		if v, ok := confMap[key]; ok {
			results[key] = v
		} else {
			fmt.Println(key, "not in config.json")
			results[key] = ""
		}
	}
	return results //return map of key-values
}

// Struct define a rule in rules.json
type RulesJSON struct {
	Comment     string   `json:"comment"`
	RuleID      string   `json:"rule-id"`
	Path        string   `json:"path"`
	Prefix      string   `json:"prefix"`
	IndexValues []string `json:"index_values"`
	Fields      []string `json:"fields"`
}

// create InfluxDB client Global variable InfluxClient
func InfluxCreateClient(tand_host string, tand_port string) {
	// Make Influx client
	url := "http://" + tand_host + ":" + tand_port
	InfluxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: url,
	})
	if err != nil {
		log.Errorln("Error creating InfluxDB Client: ", err.Error())
	}
	defer InfluxClient.Close()
	log.Infoln("InfluxDB Client connection", InfluxClient)
}

// Create the point with data for writing
func (dev *Device_Details) AddPoint(fields map[string]interface{}, time time.Time, sensor *Sensor, batchSize int) {
	// Create a point and add to batch
	tags := map[string]string{}
	pt, err := client.NewPoint(sensor.Measurement, tags, fields, time)
	if err != nil {
		log.Errorf("Device %s Create point error: %s\n", dev.DeviceName, err.Error())
	}
	// Add InfluxDB point to slice
	dev.Points = append(dev.Points, pt)
	// Check size of slice/list of Points
	if len(dev.Points) > batchSize {
		//Flush the data points in the device
		dev.FlushPoints()
	}
}

// Create Influx BatchPoints for database/device And Write datapoints from device_details list
func (dev *Device_Details) FlushPoints() {
	// Create BatchPoint
	batchPoint, error := client.NewBatchPoints(client.BatchPointsConfig{
		Database: dev.Database, //Use database from devce_details
	})
	if error != nil {
		log.Errorf("Device %s Create BatchPoint error: %s\n", dev.DeviceName, error.Error())
	}
	fmt.Printf("ERROR: %+v\n", error)
	fmt.Printf("BATCHPOINT: %+v\n", batchPoint)
	pts := dev.Points
	batchPoint.AddPoints(pts)
	fmt.Printf("INFLUXCLIENT: %+v\n", InfluxClient)
	if InfluxClient != nil {
		err := InfluxClient.Write(batchPoint) //Write batchpoint to Influx
		if err != nil {
			log.Errorf("Write Batchpoint to Influx database %s error %s\n", dev.Database, err.Error())
		} else {
			log.Infof("Write Batchpoint to Influx for %s using database: %s\n", dev.DeviceName, dev.Database)
			dev.Points = nil //Clear slice of Points in device_details back to zero
		}
	} else {
		log.Errorf("No Influx client to write data points\n")
	}
	if InfluxClient == nil {
		fmt.Printf("INFLUXCLIENT NILNILNIL: %+v\n", InfluxClient)
	}
}
