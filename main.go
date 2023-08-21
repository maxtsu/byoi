package main

import (
	"byoi/gnfingest"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gologme/log"
)

func main() {

	fmt.Println("Device-Keys ")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	//run := false
	for run {
		fmt.Printf("waiting for kafka message\n")
		time.Sleep(2 * time.Second)
		select {
		case sig := <-sigchan:
			log.Warnf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			fmt.Println("Repeating message")
		}
	}

	// Run sample files in home lab
	//hometest()

}

// Process the raw kafka message pointer to message (we do not change it)
func ProcessKafkaMessage(message *gnfingest.Message, devices []gnfingest.Device_Keys) {
	//Extract source IP and Path from message
	messageSource := message.MessageSource()
	messagePath := message.MessagePath()
	log.Debugln("Source %s Path %s", messageSource, messagePath)

	//Start matching message to configured rules in config.json
	for _, d := range devices {
		fmt.Println("Device: ", d.DeviceName)
		if (d.DeviceName == message.Source) && (d.KVS_path == message.Updates[0].Path) {
			fmt.Printf("name and prefix match ")
		}
	}

}
