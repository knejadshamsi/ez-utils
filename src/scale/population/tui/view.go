package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the complete TUI
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}
	
	// Build the complete view
	header := m.renderHeader()
	systemMetrics := m.renderSystemMetrics()
	phaseInfo := m.renderPhaseInfo()
	agentSections := m.renderAgentSections()
	footer := m.renderFooter()
	
	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		systemMetrics,
		phaseInfo,
		agentSections,
		footer,
	)
	
	return m.styles.mainContainer.Render(content)
}

// renderHeader renders the title and elapsed time
func (m Model) renderHeader() string {
	title := lipgloss.JoinHorizontal(
		lipgloss.Center,
		m.styles.titleEZ.Render("EZ"),
		m.styles.titleText.Render("Population Scaling"),
		m.styles.phaseInfo.Render(fmt.Sprintf("| %s", m.currentPhase)),
		m.FormatElapsedTime(),
	)
	
	return m.styles.sectionContainer.Render(title)
}

// renderSystemMetrics renders CPU, RAM, and disk usage
func (m Model) renderSystemMetrics() string {
	cpu := m.styles.FormatSystemMetric("CPU", fmt.Sprintf("%.1f%%", m.cpuUsage), m.styles.cpuInfo)
	ram := m.styles.FormatSystemMetric("RAM", fmt.Sprintf("%.1f%%", m.ramUsage), m.styles.ramInfo)
	disk := m.styles.FormatSystemMetric("DISK", fmt.Sprintf("%.1f MB/s", m.diskUsage), m.styles.diskInfo)
	
	metrics := lipgloss.JoinHorizontal(
		lipgloss.Left,
		cpu, "  ", ram, "  ", disk,
	)
	
	return m.styles.sectionContainer.Render(metrics)
}

// renderPhaseInfo renders phase-specific information
func (m Model) renderPhaseInfo() string {
	var info strings.Builder
	
	// Input file information
	if m.inputFile != "" {
		fileName := filepath.Base(m.inputFile)
		info.WriteString(m.styles.FormatCounter("Input File", fileName))
		info.WriteString("  ")
		info.WriteString(m.styles.FormatCounter("Size", fmt.Sprintf("%.1f MB", m.inputFileSizeMB)))
		info.WriteString("\n")
	}
	
	// Grid information
	if m.gridHandler.BinCount > 0 {
		gridDims := fmt.Sprintf("%.0fx%.0f", 
			m.gridBounds.MaxX-m.gridBounds.MinX, 
			m.gridBounds.MaxY-m.gridBounds.MinY)
		info.WriteString(m.styles.FormatCounter("Grid", gridDims))
		info.WriteString("  ")
		info.WriteString(m.styles.FormatCounter("Bins", m.gridHandler.BinCount))
		info.WriteString("  ")
		info.WriteString(m.styles.FormatCounter("Density Map", m.densityMapSize))
		info.WriteString("  ")
		info.WriteString(m.styles.FormatCounter("Safe Points", m.safePointsCount))
	}
	
	return m.styles.sectionContainer.Render(info.String())
}

// renderAgentSections renders all agent information
func (m Model) renderAgentSections() string {
	sections := []string{
		m.renderReaderAgent(),
		m.renderExtractorAgents(),
		m.renderHashMapAgents(),
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderReaderAgent renders the reader agent status
func (m Model) renderReaderAgent() string {
	var content strings.Builder
	
	content.WriteString(m.styles.systemLabel.Render("Reader Agent"))
	content.WriteString("\n")
	
	if m.readerAgent.ID != "" {
		stateStyle := m.styles.GetAgentStateStyle(m.GetAgentStateString(m.readerAgent.State))
		content.WriteString(fmt.Sprintf("  Status: %s", stateStyle.Render(m.GetAgentStateString(m.readerAgent.State))))
		content.WriteString(fmt.Sprintf("  Lines: %s", m.styles.counterInfo.Render(formatNumber(m.readerAgent.TotalLinesRead))))
		content.WriteString(fmt.Sprintf("  Persons: %s", m.styles.counterInfo.Render(formatNumber(m.readerAgent.TotalPersonsFound))))
		content.WriteString(fmt.Sprintf("  Data: %s MB", m.styles.counterInfo.Render(fmt.Sprintf("%.1f", m.readerAgent.DataProcessedMB))))
		content.WriteString(fmt.Sprintf("  Safe Points: %s", m.styles.counterInfo.Render(formatNumber(m.readerAgent.SafePointsWritten))))
	} else {
		content.WriteString("  " + m.styles.warningText.Render("Not started"))
	}
	
	return m.styles.sectionContainer.Render(content.String())
}

// renderExtractorAgents renders all extractor agents
func (m Model) renderExtractorAgents() string {
	var content strings.Builder
	
	activeCount := m.GetActiveExtractorCount()
	totalCount := m.GetTotalExtractorCount()
	
	content.WriteString(m.styles.systemLabel.Render(fmt.Sprintf("Extractor Agents (%d/%d)", activeCount, totalCount)))
	content.WriteString("\n")
	
	if totalCount > 0 {
		// Summary statistics
		totalProcessed := m.GetTotalPersonsProcessed()
		totalSkipped := m.GetTotalPersonsSkipped()
		
		content.WriteString(fmt.Sprintf("  Status: %s", m.styles.successText.Render("ACTIVE")))
		content.WriteString(fmt.Sprintf("  Processed: %s", m.styles.counterInfo.Render(formatNumber(totalProcessed))))
		content.WriteString(fmt.Sprintf("  Skipped: %s", m.styles.counterInfo.Render(formatNumber(totalSkipped))))
		content.WriteString("\n")
		
		// Individual agent details (if space allows)
		if len(m.extractorAgents) <= 5 {
			for id, agent := range m.extractorAgents {
				stateStyle := m.styles.GetAgentStateStyle(m.GetAgentStateString(agent.State))
				content.WriteString(fmt.Sprintf("  %s: %s  Processed: %s  Skipped: %s\n",
					id,
					stateStyle.Render(m.GetAgentStateString(agent.State)),
					m.styles.counterInfo.Render(formatNumber(agent.PersonsProcessed)),
					m.styles.counterInfo.Render(formatNumber(agent.PersonsSkipped))))
			}
		}
	} else {
		content.WriteString("  " + m.styles.warningText.Render("Not started"))
	}
	
	return m.styles.sectionContainer.Render(content.String())
}

// renderHashMapAgents renders all hashmap agents
func (m Model) renderHashMapAgents() string {
	var content strings.Builder
	
	activeCount := m.GetActiveHashMapCount()
	totalCount := m.GetTotalHashMapCount()
	
	content.WriteString(m.styles.systemLabel.Render(fmt.Sprintf("HashMap Agents (%d/%d)", activeCount, totalCount)))
	content.WriteString("\n")
	
	if totalCount > 0 {
		// Summary statistics
		totalCoords := m.GetTotalCoordinatesProcessed()
		totalExpansions := m.GetTotalExpansionRequests()
		avgQueueFill := m.GetAverageQueueFill()
		
		queueStyle := m.styles.counterInfo
		if avgQueueFill > 80 {
			queueStyle = m.styles.warningText
		} else if avgQueueFill > 95 {
			queueStyle = m.styles.errorText
		}
		
		content.WriteString(fmt.Sprintf("  Status: %s", m.styles.successText.Render("ACTIVE")))
		content.WriteString(fmt.Sprintf("  Queue: %s", queueStyle.Render(fmt.Sprintf("%.1f%%", avgQueueFill))))
		content.WriteString(fmt.Sprintf("  Coords: %s", m.styles.counterInfo.Render(formatNumber(totalCoords))))
		content.WriteString(fmt.Sprintf("  Expansions: %s", m.styles.counterInfo.Render(formatNumber(totalExpansions))))
		content.WriteString("\n")
		
		// Individual agent details (if space allows)
		if len(m.hashmapAgents) <= 5 {
			for id, agent := range m.hashmapAgents {
				stateStyle := m.styles.GetAgentStateStyle(m.GetAgentStateString(agent.State))
				queueStyle := m.styles.counterInfo
				if agent.QueueFillPercentage > 80 {
					queueStyle = m.styles.warningText
				}
				
				content.WriteString(fmt.Sprintf("  %s: %s  Queue: %s  Coords: %s  Exp: %s\n",
					id,
					stateStyle.Render(m.GetAgentStateString(agent.State)),
					queueStyle.Render(fmt.Sprintf("%.0f%%", agent.QueueFillPercentage)),
					m.styles.counterInfo.Render(formatNumber(agent.CoordinatesProcessed)),
					m.styles.counterInfo.Render(formatNumber(agent.ExpansionRequests))))
			}
		}
	} else {
		content.WriteString("  " + m.styles.warningText.Render("Not started"))
	}
	
	return m.styles.sectionContainer.Render(content.String())
}

// renderFooter renders control instructions
func (m Model) renderFooter() string {
	controls := []string{
		m.styles.systemLabel.Render("Controls:"),
		"q/Ctrl+C: Quit",
		"p: Pause/Resume",
		"r: Refresh",
	}
	
	footer := lipgloss.JoinHorizontal(lipgloss.Left, controls...)
	return m.styles.sectionContainer.Render(footer)
}

// Helper function to render progress bars
func (m Model) renderProgressBar(current, total int64, width int) string {
	if total == 0 {
		return m.styles.progressEmpty.Render(strings.Repeat(" ", width))
	}
	
	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))
	empty := width - filled
	
	filledBar := m.styles.progressFilled.Render(strings.Repeat("█", filled))
	emptyBar := m.styles.progressEmpty.Render(strings.Repeat("░", empty))
	
	return m.styles.progressBar.Render(filledBar + emptyBar)
}