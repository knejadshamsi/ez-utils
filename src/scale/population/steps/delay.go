package steps

import (
	"time"
)

// DelayBetweenSteps adds an artificial delay between steps to allow observation of display state
// This should be called at appropriate points in the workflow
func DelayBetweenSteps(seconds int) {
	// Simple sleep for the specified number of seconds
	time.Sleep(time.Duration(seconds) * time.Second)
}
