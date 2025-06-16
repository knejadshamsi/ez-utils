package display

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case stepMsg:
		// Store step durations
		if m.stepNumber >= 0 && int(msg) != m.stepNumber {
			if m.stepDurations == nil {
				m.stepDurations = make(map[int]time.Duration)
			}
			m.stepDurations[m.stepNumber] = m.elapsedTime
			m.timer = timer.NewWithInterval(0, time.Second)
			m.elapsedTime = 0
			m.startTime = time.Now()
		}

		m.stepNumber = int(msg)
		return m, nil

	case processCompleteMsg:
		m.processComplete = bool(msg)
		return m, nil

	// Handlers are organized by processing phase: XML(0-1) -> Location(2-3) -> Processing(4-18) -> System
	case chunkCountMsg:
		m.chunkCount = int(msg)
		return m, nil

	case personsFoundMsg:
		m.personsFound = int(msg)
		return m, nil

	case bytesReadMsg:
		m.bytesReadMB = int(msg)
		return m, nil

	case firstChunkFixedMsg:
		m.firstChunkFixed = bool(msg)
		return m, nil

	case lastChunkFixedMsg:
		m.lastChunkFixed = bool(msg)
		return m, nil

	case agentCounterMsg:
		m.agentCounter = int(msg)
		return m, nil

	case chunkCounterMsg:
		m.chunkCounter.Current = msg.current
		m.chunkCounter.Total = msg.total
		return m, nil

	case dbCounterMsg:
		m.dbCounter.Current = msg.current
		m.dbCounter.Total = msg.total
		return m, nil

	case coordinateCounterMsg:
		m.coordinateCount = int(msg)
		return m, nil

	case binCounterMsg:
		m.binCounter.Current = msg.current
		m.binCounter.Total = msg.total
		return m, nil

	case agentBinCounterMsg:
		m.agentBinCount = int(msg)
		return m, nil

	case scaleCounterMsg:
		m.scaleCounter.Current = msg.current
		m.scaleCounter.Total = msg.total
		return m, nil

	case outputScaleMsg:
		m.outputScale = int(msg)
		return m, nil

	case outputCounterMsg:
		m.outputCounter.Current = msg.current
		m.outputCounter.Total = msg.total
		return m, nil

	case cleanupCounterMsg:
		m.cleanupFiles = msg.files
		m.cleanupDirs = msg.dirs
		m.cleanupBytes = msg.bytes
		return m, nil

	case systemStatsMsg:
		m.cpuUsage = msg.cpu
		m.ramUsage = msg.ram
		m.diskUsage = msg.disk
		return m, nil

	case tea.WindowSizeMsg:
		// TODO: We can handle window size changes here if needed in the future
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case timer.TimeoutMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)

	case tickMsg:
		m.elapsedTime = time.Since(m.startTime)
		m.totalElapsedTime = time.Since(m.totalStartTime)
		cmds = append(cmds, tickCmd())

	case blinkMsg:
		m.statusBlink = !m.statusBlink
		cmds = append(cmds, blinkCmd())
	}

	// Always add the spinner.Tick command to ensure the spinner keeps animating
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}

	// If no commands were added, still return some to keep animations going
	return m, tea.Batch(spinner.Tick)
}

func parseIntWithDefault(s string, defaultVal int) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultVal, err
	}
	return result, nil
}
