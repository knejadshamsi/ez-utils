package core

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"

	"ez-utils/src/display/tui"
)

// monitorSystem monitors CPU and RAM usage for the display instance
func (d *DisplayInstance) monitorSystem() {
	// Create ticker for more frequent updates during short processes
	ticker := time.NewTicker(500 * time.Millisecond) // More frequent for short-running processes
	stopChan := make(chan bool, 1)
	
	// Add cleanup function to stop monitoring
	d.AddCleanupFunc(func() {
		ticker.Stop()
		select {
		case stopChan <- true:
		default:
		}
	})

	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			// Check if still running before doing expensive operations
			d.mutex.RLock()
			isRunning := d.running
			program := d.program
			d.mutex.RUnlock()

			if !isRunning || program == nil {
				return
			}

			// Get system stats with timeout
			cpuPercent, err := cpu.Percent(500*time.Millisecond, false)
			cpuUsage := 0.0
			if err == nil && len(cpuPercent) > 0 {
				cpuUsage = cpuPercent[0]
			}

			memInfo, err := mem.VirtualMemory()
			ramUsage := 0.0
			if err == nil {
				ramUsage = float64(memInfo.Used) / 1024 / 1024 // Convert to MB
			}

			// Send update only if instance is still running
			d.mutex.RLock()
			if d.running && d.program != nil {
				d.program.Send(tui.SystemStatsMsg{
					Cpu:  cpuUsage,
					Ram:  ramUsage,
				})
			}
			d.mutex.RUnlock()
		}
	}
}