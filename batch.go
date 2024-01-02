package main

import (
	"fmt"
	"strconv"
	"time"
)

const size = 20     //batch size
const period = 8000 // flush timer ms

// main function
func main() {
	g1 := create("one")
	g2 := create("two")
	device_detail := []*GroupData{g1, g2}

	//create map of items
	//var device_details = make(map[string]*GroupData)
	//	device_details[g1.Name] = g1
	//	device_details[g2.Name] = g2
	var device_details = make(map[string]*GroupData)
	for _, d := range device_detail {
		device_details[d.Name] = d
	}

	for i := 1; i < 100; i++ {
		fmt.Printf("For loop %+v\n", i)
		g1.print()
		g2.print()

		message := "String" + strconv.Itoa(i)
		//g1.Point <- message
		addPoint1(message, g1)

		time.Sleep(1 * time.Second)
	}
}

type GroupData struct {
	Name   string
	Timer  *time.Timer
	Points []string
	Point  chan string
}

func (g *GroupData) print() {
	fmt.Printf("%+v %+v\n", g.Name, g.Points)
}

func create(name string) *GroupData {
	var g GroupData
	g.Name = name

	g.Timer = time.NewTimer(period * time.Millisecond)
	g.Point = make(chan string)
	go g.timer()
	fmt.Printf("Created object %+v\n", g)
	//return &g
	return &g
}

func (g *GroupData) timer() {
	for {
		select {
		case <-g.Timer.C:
			fmt.Printf("Timer %s fired\n", g.Name)
			g.Timer.Stop()
			go g.flush()
			g.Timer = time.NewTimer(period * time.Millisecond)

		case point := <-g.Point:
			fmt.Printf("Received point %s\n", point)
			g.Points = append(g.Points, point)
			if len(g.Points) > size {
				fmt.Printf("Max batch size flush: %s\n", g.Name)
				g.Timer.Stop()
				go g.flush()
				g.Timer = time.NewTimer(period * time.Millisecond)

			}
		}
	}
}

func addPoint2(pt string, g *GroupData) {
	g.Point <- pt
}

func addPoint1(pt string, g *GroupData) {
	addPoint2(pt, g)
}

func (g *GroupData) flush() {
	fmt.Printf("Flush function for %+v : %+v\n", g.Name, g.Points)
	g.Points = nil
}
