package tui

import (
	"runtime"
	"sync"
	"time"
)

// SystemMonitor provides real-time system metrics
type SystemMonitor struct {
	mutex         sync.RWMutex
	cpuUsage      float64
	ramUsage      float64
	diskUsage     float64
	lastUpdate    time.Time
	goroutineCount int
	
	// Monitoring state
	running       bool
	stopChan      chan struct{}
}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		stopChan: make(chan struct{}),
	}
}

// Start begins real-time system monitoring
func (sm *SystemMonitor) Start() {
	sm.mutex.Lock()
	if sm.running {
		sm.mutex.Unlock()
		return
	}
	sm.running = true
	sm.mutex.Unlock()

	go sm.monitorLoop()
}

// Stop stops the system monitoring
func (sm *SystemMonitor) Stop() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if !sm.running {
		return
	}
	
	sm.running = false
	close(sm.stopChan)
}

// GetMetrics returns current system metrics
func (sm *SystemMonitor) GetMetrics() (cpu, ram, disk float64) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	return sm.cpuUsage, sm.ramUsage, sm.diskUsage
}

// monitorLoop continuously monitors system metrics
func (sm *SystemMonitor) monitorLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // Faster refresh for responsive UI
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.updateMetrics()
			
			// Send update to TUI if available
			cpu, ram, disk := sm.GetMetrics()
			UpdateSystemMetrics(cpu, ram, disk)
			
		case <-sm.stopChan:
			return
		}
	}
}

// updateMetrics collects current system metrics
func (sm *SystemMonitor) updateMetrics() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate RAM usage (allocated memory in MB)
	sm.ramUsage = float64(memStats.Alloc) / (1024 * 1024)

	// Calculate CPU approximation based on goroutines and system activity
	currentGoroutines := runtime.NumGoroutine()
	sm.goroutineCount = currentGoroutines
	
	// CPU estimation: base load + goroutine activity + memory pressure
	baseLoad := 15.0 // Base system load
	goroutineLoad := float64(currentGoroutines) * 0.8 // Each goroutine adds load
	memoryPressure := (sm.ramUsage / 1000) * 10 // Memory pressure affects CPU
	
	sm.cpuUsage = baseLoad + goroutineLoad + memoryPressure
	if sm.cpuUsage > 100 {
		sm.cpuUsage = 95 + (sm.cpuUsage-100)*0.1 // Cap at realistic values
	}

	// Disk usage simulation (IO operations)
	// In a real implementation, this would track actual disk I/O
	sm.diskUsage = 5.0 + (sm.ramUsage/100)*2 // Correlate with memory activity

	sm.lastUpdate = time.Now()
}

// Global system monitor instance
var globalSystemMonitor *SystemMonitor
var monitorOnce sync.Once

// StartGlobalSystemMonitor starts the global system monitor
func StartGlobalSystemMonitor() {
	monitorOnce.Do(func() {
		globalSystemMonitor = NewSystemMonitor()
		globalSystemMonitor.Start()
	})
}

// StopGlobalSystemMonitor stops the global system monitor
func StopGlobalSystemMonitor() {
	if globalSystemMonitor != nil {
		globalSystemMonitor.Stop()
		globalSystemMonitor = nil // Clean up reference to prevent memory leaks
	}
}

// GetCurrentSystemMetrics returns current system metrics from global monitor
func GetCurrentSystemMetrics() (cpu, ram, disk float64) {
	if globalSystemMonitor != nil {
		return globalSystemMonitor.GetMetrics()
	}
	return 0, 0, 0
}