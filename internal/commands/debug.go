package commands

import "fmt"

var debugMode bool

func SetDebugMode(enabled bool) {
	debugMode = enabled
}

func debugf(format string, args ...interface{}) {
	if debugMode {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}
