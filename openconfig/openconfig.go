package openconfig

// Struct for Message from gnmic
type Message struct {
	Source           string    `json:"source"`
	SubscriptionName string    `json:"subscription-name"`
	Timestamp        int64     `json:"timestamp"`
	Time             string    `json:"time"`
	Prefix           string    `json:"prefix"`
	Updates          []Updates `json:"updates"`
}

// Updates level
type Updates struct {
	Path   string `json:"Path"`
	Values Values `json:"values"`
}

type Values struct {
	InterfacesInterfaceStateCounters struct {
		InBroadcastPkts  int64 `json:"in-broadcast-pkts"`
		InDiscards       int64 `json:"in-discards"`
		InErrors         int64 `json:"in-errors"`
		InFcsErrors      int64 `json:"in-fcs-errors"`
		InMulticastPkts  int64 `json:"in-multicast-pkts"`
		InOctets         int64 `json:"in-octets"`
		InPkts           int64 `json:"in-pkts"`
		InUnicastPkts    int64 `json:"in-unicast-pkts"`
		InUnknownProtos  int64 `json:"in-unknown-protos"`
		OutBroadcastPkts int64 `json:"out-broadcast-pkts"`
		OutDiscards      int64 `json:"out-discards"`
		OutErrors        int64 `json:"out-errors"`
		OutMulticastPkts int64 `json:"out-multicast-pkts"`
		OutOctets        int64 `json:"out-octets"`
		OutPkts          int64 `json:"out-pkts"`
		OutUnicastPkts   int64 `json:"out-unicast-pkts"`
	} `json:"interfaces/interface/state/counters"`
}
