// package gnfingest
package openconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// interface
type I interface {
	getValue(key string) interface{}
}

// gnmic Event Message partial struct
type Message struct {
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
	Tags      struct {
		Path             string `json:"path"`
		Prefix           string `json:"prefix"`
		Source           string `json:"source"`
		SubscriptionName string `json:"subscription-name"`
	} `json:"tags"`
	Values json.RawMessage `json:"values"`
}

// Message Method to verify openconfig JSON event message
func (m *Message) MessageVerify() error {
	if m.Name == "" {
		return fmt.Errorf("no Name field in message")
	} else if m.Tags.Prefix == "" {
		return fmt.Errorf("no Prefix field in message")
	} else if m.Tags.Source == "" {
		return fmt.Errorf("no Source field in message")
	} else if string(m.Tags.Path) == "" {
		return fmt.Errorf("no Path field in message")
	} else if string(m.Values) == "" {
		return fmt.Errorf("no Values field in message")
	} else { // This openconfig message is OK
		return nil
	}
}

// Message Method to extract source IP & path
func (m *Message) MessageSource() string {
	// Extract message source IP remove port number
	var source string
	if strings.Contains(m.Tags.Source, ":") {
		s := strings.Split(m.Tags.Source, ":")
		source = s[0]
	} else {
		source = m.Tags.Source
	}
	return source
}

// gnmic Message partial struct
type OldMessage struct {
	Source           string `json:"source"`
	SubscriptionName string `json:"subscription-name"`
	Timestamp        int64  `json:"timestamp"`
	Time             string `json:"time"`
	Prefix           string `json:"prefix"`
	Updates          []struct {
		Path   string          `json:"Path"`
		Values json.RawMessage `json:"values"`
	}
}

// Message Method to verify openconfig JSON message
func (m *OldMessage) MessageVerify() error {
	if m.Source == "" {
		return fmt.Errorf("no Source field in message")
	} else if m.Prefix == "" {
		return fmt.Errorf("no Prefix field in message")
	} else if len(m.Updates) <= 0 {
		return fmt.Errorf("no updates array in message")
	} else if m.Updates[0].Path == "" {
		return fmt.Errorf("no path field in message")
	} else { // This openconfig message is OK
		return nil
	}
}

// Message Method to extract source IP & path
func (m *OldMessage) MessageSource() string {
	// Extract message source IP remove port number
	return (strings.Split(m.Source, ":")[0])
}

// Message Method to extract message path
func (m *OldMessage) MessagePath() string {
	// Extract message path remove index values []
	re := regexp.MustCompile("[[].*?[]]")
	return (re.ReplaceAllString(m.Updates[0].Path, ""))
}

// oc-interfaces Values with different paths struct
type Values struct {
	Counters *Counters `json:"interfaces/interface/state/counters,omitempty"`
	State    *State    `json:"interfaces/interface/state,omitempty"`
	Isis     *Isis     `json:"network-instances/network-instance/protocols/protocol/isis,omitempty"`
}

// oc-interfaces Values with interfaces/interface/state
type OpenconfigInterface struct {
	State State `json:"interfaces/interface/state"`
}

// Top level entry point for get fields from child structs
// openconfig-interfaces:
// pass list of paths fields (but only from the first slice of paths)
func (s *OpenconfigInterface) GetFields(path []string, value []string) map[string]string {
	result := map[string]string{} //where the fields results go
	i := len(path)
	pathFirst := path[0] //take first item of path and get the next struct
	if i > 1 {           //get path from  next child struct
		path2 := path[1:]      //pop/remove first item in path list
		path2First := path2[0] //first item of new list
		switch path2First {
		default:
			fmt.Printf("error ")
		}
	} else { //getting values from the child struct
		switch pathFirst {
		case "interfaces/interface/state":
			childStruct := s.State
			result = getValueI(&childStruct, value)
		}

	}
	return result
}

// InterfacesInterfaceState method to return path and values as selected
func (s *OpenconfigInterface) getPath(path []string, values []string) []string {
	i := len(path)
	p := path[0] //take first item of path and get the next struct
	result := []string{}
	switch p {
	case "interfaces/interface/state": //no path check as it is final there is no child structs
		fmt.Printf("enter into state struct\n")
		//childStruct := s.Test
		//result = getValueI(&childStruct, values)
	default:
		result = nil
	}
	fmt.Printf("i %s result %+v\n", i, result)
	return result
}

// oc-interfaces State struct
type State struct {
	AdminStatus string `json:"admin-status"`
	//Counters     Counters `json:"counters"`
	Enabled      bool   `json:"enabled"`
	Ifindex      int    `json:"ifindex"`
	LastChange   string `json:"last-change"`
	Logical      bool   `json:"logical"`
	LoopbackMode bool   `json:"loopback-mode"`
	Mtu          int    `json:"mtu"`
	Name         string `json:"name"`
	OperStatus   string `json:"oper-status"`
	Type         string `json:"type"`
}

// State method to return values as selected
func (s *State) getValue(key string) interface{} {
	switch key {
	case "admin-status":
		return s.AdminStatus
	/*case "counters":
	return s.Counters */
	case "enabled":
		return s.Enabled
	case "ifindex":
		return s.Ifindex
	case "last-change":
		return s.LastChange
	case "logical":
		return s.Logical
	case "loopback-mode":
		return s.LoopbackMode
	case "mtu":
		return s.Mtu
	case "name":
		return s.Name
	case "oper-status":
		return s.OperStatus
	case "type":
		return s.Type
	default:
		return nil
	}
}

// State method to return path and values as selected
func (s *State) getPath(path []string, values []string) []string {
	i := len(path)
	p := path[0] //take first item of path and get the next struct
	result := []string{}
	switch p {
	case "test": //no path check as it is final there is no child structs
		fmt.Printf("enter into TEST struct\n")
		//childStruct := s.Test
		//result = getValueI(&childStruct, values)
	default:
		result = nil
	}
	fmt.Printf("i %s result %+v\n", i, result)
	return result
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

// getValue using interface on the struct
func getValueI(msg I, keys []string) map[string]string {
	if keys != nil {
		values := map[string]string{}
		for _, v := range keys {
			tmp := msg.getValue(v)
			result := fmt.Sprintf("%s", tmp)
			values[v] = result //create a map of value fields
		}
		fmt.Printf("tpye of values: %s\n", reflect.TypeOf(values))
		return values
	} else {
		return nil //not required to return values
	}
}
