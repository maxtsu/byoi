package hbinconfig

// The Hbin struct for the config.json
type Hbin struct {
	Inputs  []Inputs  `json:"inputs"`
	Outputs []Outputs `json:"outputs"`
}

// The Outputs struct for the config.json
type Outputs struct {
	OutPlugin []OutPlugin `json:"plugin"`
}

// The Output Plugin struct for the config.json
type OutPlugin struct {
	Name      string      `json:"name"`
	OutConfig []OutConfig `json:"config"`
}

// The Config struct for the config.json
type OutConfig struct {
	Server   string `json:"server"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// The Inputs struct for the config.json
type Inputs struct {
	Plugin []Plugin `json:"plugin"`
}

// The Plugin struct for the config.json
type Plugin struct {
	Name   string   `json:"name"`
	Config []Config `json:"config"`
}

// The Config struct for the config.json
type Config struct {
	Device []Device `json:"device"`
	Devgrp string   `json:"device-group"`
	KVs    []KVs    `json:"kvs"`
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
