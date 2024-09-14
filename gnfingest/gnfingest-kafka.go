package gnfingest

import (
	"strings"
)

// Check kafka message key against list of devices
func KafkaKeyCheck(key []byte) {
	stringKey := string(key)
	ip := strings.Split(stringKey, ":")
}
