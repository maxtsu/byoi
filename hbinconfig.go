package hbinconfig

import "log"

// The Config struct for the config.json
type Config struct {
	Device   []Device `json:"device"`
	Security Security `json:"security"`
	KVs      []KVs    `json:"kvs"`
	logger   log.Logger
}

// Device type
type Device struct {
	Name      string         `json:"name"`
	SystemID  string         `json:"system-id"`
	Sensor    []Sensor       `json:"sensor"`
	HBStorage HBStorage      `json:"healthbot-storage"`
	Auth      Authentication `json:"authentication"`
}

// Sensor config
type Sensor struct {
	Name        string `json:"name"`
	KVs         []KVs  `json:"kvs"`
	Measurement string `json:"measurement"`
}

// HBStorage to specify DB name and retention policy name for this device
type HBStorage struct {
	DB              string `json:"database"`
	RetentionPolicy string `json:"retention-policy"`
}

// Authentication config
type Authentication struct {
}

// Security config
type Security struct {
}

// KVs config
type KVs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
