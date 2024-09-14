package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var group1 = (os.Getenv("TERM'") + "-golang1")
var group2 = (os.Getenv("TERM") + "-golang1")

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(requestDuration)
}

func handler(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues(r.Method, r.URL.Path))
	defer timer.ObserveDuration()
	httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

	// Your handler logic
}

func stocks() {
	var stocks map[string]float64
	sym := "TTWO"
	price := stocks[sym]
	fmt.Printf("1. %s -> $%.2f\n", sym, price)

	if price, ok := stocks[sym]; ok { // using comma ok to check if key exists
		fmt.Printf("2. %s -> $%.2f\n", sym, price)
	} else {
		fmt.Printf("2. %s not found\n", sym)
	}

	stocks = make(map[string]float64) // It's importatnt to initialize the map before adding values. Else it's a panic situation
	stocks[sym] = 123.4
	stocks["APPL"] = 345.6

	if price, ok := stocks[sym]; ok {
		fmt.Printf("3. %s -> $%.2f\n", sym, price)
	} else {
		fmt.Printf("3. %s not found\n", sym)
	}
}

// main function
func main() {
	// http.Handle("/metrics", promhttp.Handler())
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)

	fmt.Printf("group1 = %+v\n", group1)
	fmt.Printf("group2 = %+v\n", group2)
	var fields string
	//fields = "interfaces/interface/state/oper-status,interfaces/interface/state/admin-status"
	fields = ""
	// Split the field string by comma
	f1 := strings.Split(fields, ",")

	var Fields []string
	for _, f2 := range f1 {
		// Trim whitespace from the fields
		s := strings.TrimSpace(f2)
		Fields = append(Fields, s)
	}

	//run := true
	//if !run {
	//	fmt.Println("Exiting")
	//	os.Exit(1)
	//}

	stocks()

	// shell()
	// fmt.Printf("result %+v\n", Fields)

}

func shell() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		fmt.Println("typed text: ", text)

	}
}
