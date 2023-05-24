package gnfingest

// gnmic Message struct
type Message struct {
	Source           string `json:"source"`
	SubscriptionName string `json:"subscription-name"`
	Timestamp        int64  `json:"timestamp"`
	Time             string `json:"time"`
	Prefix           string `json:"prefix"`
	Updates          []struct {
		Path   string `json:"Path"`
		Values Values `json:"values"`
	}
}

// oc-interfaces Values with different paths struct
type Values struct {
	Counters Counters `json:"interfaces/interface/state/counters"`
	State    State    `json:"interfaces/interface/state"`
	Isis     Isis     `json:"network-instances/network-instance/protocols/protocol/isis"`
}

// oc-interfaces State struct
type State struct {
	AdminStatus  string   `json:"admin-status"`
	Counters     Counters `json:"counters"`
	Enabled      bool     `json:"enabled"`
	Ifindex      int      `json:"ifindex"`
	LastChange   string   `json:"last-change"`
	Logical      bool     `json:"logical"`
	LoopbackMode bool     `json:"loopback-mode"`
	Mtu          int      `json:"mtu"`
	Name         string   `json:"name"`
	OperStatus   string   `json:"oper-status"`
	Type         string   `json:"type"`
}

// oc-interfaces Counters struct
type Counters struct {
	InBroadcastPkts    string `json:"in-broadcast-pkts"`
	InDiscards         string `json:"in-discards"`
	InErrors           int    `json:"in-errors"`
	InFcsErrors        string `json:"in-fcs-errors"`
	InMulticastPkts    string `json:"in-multicast-pkts"`
	InOctets           string `json:"in-octets"`
	InPkts             string `json:"in-pkts"`
	InUnicastPkts      string `json:"in-unicast-pkts"`
	InUnknownProtos    string `json:"in-unknown-protos"`
	OutBroadcastPkts   string `json:"out-broadcast-pkts"`
	OutDiscards        string `json:"out-discards"`
	OutErrors          int    `json:"out-errors"`
	OutMulticastPkts   string `json:"out-multicast-pkts"`
	OutOctets          string `json:"out-octets"`
	OutPkts            string `json:"out-pkts"`
	OutUnicastPkts     string `json:"out-unicast-pkts"`
	CarrierTransitions string `json:"carrier-transitions"`
}

// oc-network-instance ISIS adjacency struct
type Isis struct {
	Interfaces struct {
		Interface struct {
			InterfaceID string `json:"interface-id"`
			Levels      struct {
				Level struct {
					Adjacencies struct {
						Adjacency struct {
							State struct {
								AdjacencyState            string   `json:"adjacency-state"`
								AdjacencyType             string   `json:"adjacency-type"`
								AreaAddress               []string `json:"area-address"`
								LocalExtendedCircuitID    int      `json:"local-extended-circuit-id"`
								MultiTopology             bool     `json:"multi-topology"`
								NeighborCircuitType       string   `json:"neighbor-circuit-type"`
								NeighborExtendedCircuitID int      `json:"neighbor-extended-circuit-id"`
								NeighborIpv4Address       string   `json:"neighbor-ipv4-address"`
								NeighborIpv6Address       string   `json:"neighbor-ipv6-address"`
								NeighborSnpa              string   `json:"neighbor-snpa"`
								Nlpid                     []string `json:"nlpid"`
								Priority                  int      `json:"priority"`
								RemainingHoldTime         int      `json:"remaining-hold-time"`
								RestartStatus             bool     `json:"restart-status"`
								RestartSupport            bool     `json:"restart-support"`
								RestartSuppress           bool     `json:"restart-suppress"`
								SystemID                  string   `json:"system-id"`
								Topology                  []string `json:"topology"`
								UpTime                    int      `json:"up-time"`
							} `json:"state"`
							SystemID string `json:"system-id"`
						} `json:"adjacency"`
					} `json:"adjacencies"`
					LevelNumber int `json:"level-number"`
				} `json:"level"`
			} `json:"levels"`
		} `json:"interface"`
	} `json:"interfaces"`
}
