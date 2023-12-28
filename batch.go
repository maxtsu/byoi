package main

import (
	"fmt"
	"time"
)

const size = 4      //batch size
const period = 9000 // flush timer ms

// main function
func main() {
	g1 := create("one")
	g2 := create("two")
	g2.print()
	for i := 1; i < 100; i++ {
		fmt.Printf("For loop %+v\n", i)
		g1.add(1)
		//g1.print()
		//g2.print()
		//fmt.Printf("Print inside for loop %+v\n", g1)
		//fmt.Printf("Print inside for loop %+v\n", g2)
		time.Sleep(1 * time.Second)
	}
}

type GroupData struct {
	Name          string
	Number        int
	Timeout       *time.Ticker
	TimeoutTicker chan bool
	Ping          chan struct{}
}

func (g *GroupData) print() {
	fmt.Printf("%+v %+v\n", g.Name, g.Number)
}

func create(name string) *GroupData {
	var g GroupData
	g.Name = name
	g.Timeout = time.NewTicker(period * time.Millisecond)
	gp := &g
	go gp.periodic()
	fmt.Printf("Create object %+v\n", g)
	return &g
}

func (g *GroupData) periodic() {
	signal := make(chan struct{})
	g.Ping = signal
	for {
		select {
		case <-g.Timeout.C: // Activate periodically
			fmt.Printf("Timeout:  %+v \n", g.Name)
			g.flush()
		case <-signal:
			fmt.Printf("Ping\n")
			g.Timeout.Stop()
			g.Timeout = time.NewTicker(period * time.Millisecond)
		default:
			continue
		}
	}
}

func (g *GroupData) add(n int) {
	g.Number = g.Number + n
	if g.Number == size {
		//g.flush()
		//g.Number = 0
		g.Ping <- struct{}{}
	}
}

func (g *GroupData) flush() {
	fmt.Printf("Flush function for %+v\n", g.Name)
}
