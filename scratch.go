package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// var rulesfile = "even_rules.json"
var yamlfile = "test.yaml"

// Struct for Yaml
type File []YamlRules

type YamlRules struct {
	RuleID      string   `yaml:"rule-id"`
	Path        string   `yaml:"path"`
	Prefix      string   `yaml:"prefix,omitempty"`
	IndexValues []string `yaml:"index-values"`
	Fields      []string `yaml:"fields"`
}

// main function
func main() {
	var r File

	// read the yaml file
	data, err := os.ReadFile(yamlfile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &r)
	if err != nil {
		panic(err)
	}
	fmt.Printf("yaml: %+v\n", r)

	var a = make(map[string]YamlRules)
	for k, v := range r {
		fmt.Printf("k: %+v v: %+v\n", k, v)
		fmt.Printf("Rule-ID: %+v\n", v.RuleID)
		r3 := v.RuleID
		a[r3] = v
	}
	fmt.Printf("a: %+v\n", a)
}
