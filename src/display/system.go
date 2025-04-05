package display

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func monitorSystem() {
	for {
		cpuPercent, err := cpu.Percent(time.Second, false)
		cpuUsage := 0.0
		if err == nil && len(cpuPercent) > 0 {
			cpuUsage = cpuPercent[0]
		}

		memInfo, err := mem.VirtualMemory()
		ramUsage := 0.0
		if err == nil {
			ramUsage = float64(memInfo.Used) / 1024 / 1024
		}

		if p != nil {
			p.Send(systemStatsMsg{
				cpu: cpuUsage,
				ram: ramUsage,
			})
		}

		time.Sleep(time.Second)
	}
}
