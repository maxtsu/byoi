package main

import (
	"fmt"
	"os"
)

var rulesfile = "even_rules.json"

// main function
func main() {
	if _, err := os.Stat(rulesfile); err == nil {
		fmt.Printf("File exists\n")
	} else {
		fmt.Printf("File does not exist\n")
	}
}
