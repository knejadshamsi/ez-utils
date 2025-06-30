package tui

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

	case StepMsg:
		// Store step durations
		if m.StepNumber >= 0 && int(msg) != m.StepNumber {
			if m.StepDurations == nil {
				m.StepDurations = make(map[int]time.Duration)
			}
			m.StepDurations[m.StepNumber] = m.ElapsedTime
			m.Timer = timer.NewWithInterval(0, time.Second)
			m.ElapsedTime = 0
			m.StartTime = time.Now()
		}

		m.StepNumber = int(msg)
		return m, nil

	case ProcessCompleteMsg:
		m.ProcessComplete = bool(msg)
		return m, nil

	// Handlers are organized by processing phase: XML(0-1) -> Location(2-3) -> Processing(4-18) -> System
	case ChunkCountMsg:
		m.ChunkCount = int(msg)
		return m, nil

	case PersonsFoundMsg:
		m.PersonsFound = int(msg)
		return m, nil

	case BytesReadMsg:
		m.BytesReadMB = int(msg)
		return m, nil

	case FirstChunkFixedMsg:
		m.FirstChunkFixed = bool(msg)
		return m, nil

	case LastChunkFixedMsg:
		m.LastChunkFixed = bool(msg)
		return m, nil

	case AgentCounterMsg:
		m.AgentCounter = int(msg)
		return m, nil

	case ChunkCounterMsg: // Changed to ChunkCounterMsg
		m.ChunkCounter.Current = msg.Current
		m.ChunkCounter.Total = msg.Total
		return m, nil

	case DbCounterMsg: // Changed to DbCounterMsg
		m.DbCounter.Current = msg.Current
		m.DbCounter.Total = msg.Total
		return m, nil

	case CoordinateCounterMsg: // Changed to CoordinateCounterMsg
		m.CoordinateCount = int(msg) // Changed to m.CoordinateCount
		return m, nil

	case BinCounterMsg: // Changed to BinCounterMsg
		m.BinCounter.Current = msg.Current
		m.BinCounter.Total = msg.Total
		return m, nil

	case AgentBinCounterMsg: // Changed to AgentBinCounterMsg
		m.AgentBinCount = int(msg) // Changed to m.AgentBinCount
		return m, nil

	case ScaleCounterMsg: // Changed to ScaleCounterMsg
		m.ScaleCounter.Current = msg.Current
		m.ScaleCounter.Total = msg.Total
		return m, nil

	case OutputScaleMsg: // Changed to OutputScaleMsg
		m.OutputScale = int(msg) // Changed to m.OutputScale
		return m, nil

	case OutputCounterMsg: // Changed to OutputCounterMsg
		m.OutputCounter.Current = msg.Current
		m.OutputCounter.Total = msg.Total
		return m, nil

	case CleanupCounterMsg: // Changed to CleanupCounterMsg
		m.CleanupFiles = msg.Files
		m.CleanupDirs = msg.Dirs
		m.CleanupBytes = msg.Bytes
		return m, nil

	case SystemStatsMsg: // Changed to SystemStatsMsg
		m.CPUUsage = msg.Cpu
		m.RAMUsage = msg.Ram
		// m.DiskUsage = msg.Disk // Field removed from SystemStatsMsg
		return m, nil

	case tea.WindowSizeMsg:
		// TODO: We can handle window size changes here if needed in the future
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg) // Changed to m.Spinner
		cmds = append(cmds, cmd)

	case timer.TimeoutMsg:
		var cmd tea.Cmd
		m.Timer, cmd = m.Timer.Update(msg) // Changed to m.Timer
		cmds = append(cmds, cmd)

	case TickMsg:
		m.ElapsedTime = time.Since(m.StartTime)
		m.TotalElapsedTime = time.Since(m.TotalStartTime)
		cmds = append(cmds, TickCmd())

	case BlinkMsg:
		m.StatusBlink = !m.StatusBlink
		cmds = append(cmds, BlinkCmd())
	}

	// Always add the spinner.Tick command to ensure the spinner keeps animating
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}

	// If no commands were added, still return some to keep animations going
	return m, tea.Batch(spinner.Tick)
}

func ParseIntWithDefault(s string, defaultVal int) (int, error) { // Made public
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultVal, err
	}
	return result, nil
}

