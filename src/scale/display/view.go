package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"ez-utils/src/scale/display/i18n"
)

// Handles the UI rendering
func (m Model) View() string {
	if m.processComplete {
		tm := GetTextManager()
		if tm != nil {
			return lipglossStyle.processComplete.Render(tm.GetUIText("process_complete")) + "\n"
		}
		return lipglossStyle.processComplete.Render("Process completed successfully!") + "\n"
	}

	titleSection := createTitleSection(m)
	previousStepsSection := createPreviousStepsSection(m)
	currentStepSection := createCurrentStepSection(m)
	systemInfoSection := createSystemInfoSection(m)
	upcomingStepsSection := createUpcomingStepsSection(m)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleSection,
		previousStepsSection,
		currentStepSection,
		systemInfoSection,
		upcomingStepsSection,
	)
}

func createTitleSection(m Model) string {
	tm := GetTextManager()
	if tm == nil {
		// Fallback if text manager not initialized
		ezPart := "[EZ-UTILS]"
		descPart := ": Scaling down MATSim population "
		timeStr := fmt.Sprintf("[TIME %02d:%02d]", int(m.totalElapsedTime.Minutes()), int(m.totalElapsedTime.Seconds())%60)
		abortStr := " [Press q to abort]"
		
		titleContent := lipglossStyle.titleEZ.Render(ezPart) + lipglossStyle.titleText.Render(descPart)
		rightContent := lipglossStyle.timeInfo.Render(timeStr) + lipglossStyle.titleEZ.Render(abortStr)
		return titleContent + rightContent
	}

	ezPart := tm.GetUIText("title_prefix")
	descPart := tm.GetUIText("title_description")
	timeStr := tm.FormatTimeText(int(m.totalElapsedTime.Minutes()), int(m.totalElapsedTime.Seconds())%60)
	abortStr := tm.GetUIText("abort_hint")

	titleContent := lipglossStyle.titleEZ.Render(ezPart) + lipglossStyle.titleText.Render(descPart)
	rightContent := lipglossStyle.timeInfo.Render(timeStr) + lipglossStyle.titleEZ.Render(abortStr)

	return titleContent + rightContent
}

func createPreviousStepsSection(m Model) string {
	var lines []string

	for i := 0; i < m.stepNumber; i++ {
		stepLabel := GetStepLabel(m.moduleName, i)

		completedText := fmt.Sprintf("%d. %s", i+1, stepLabel)

		var durationStr string
		if duration, ok := m.stepDurations[i]; ok {
			durationStr = formatDuration(duration)
		} else {
			durationStr = "00:00" // Default if no duration stored
		}

		completedTaskText := lipglossStyle.completedTask.Render(completedText)
		durationText := lipglossStyle.completedDuration.Render(fmt.Sprintf("{%s}", durationStr))

		completedLine := fmt.Sprintf("%s %s", completedTaskText, durationText)
		lines = append(lines, completedLine)
	}

	if len(lines) == 0 {
		return "" // No completed steps yet
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func createCurrentStepSection(m Model) string {
	tm := GetTextManager()
	var currentlyText string
	if tm != nil {
		currentlyText = tm.GetUIText("currently_prefix")
	} else {
		currentlyText = "Currently"
	}
	
	currentlyPrefix := lipglossStyle.spinner.Render(m.spinner.View()) + " " + lipglossStyle.currentlyPrefix.Render(currentlyText)

	counters := getStepCounters(m)

	var stepDetails string

	// Special case: First 4 steps have counter text that includes step number and description
	if m.stepNumber == 0 || m.stepNumber == 1 || m.stepNumber == 2 || m.stepNumber == 3 {
		stepDetails = counters
	} else {
		// For other steps, add the numbered step label plus the counters
		currentStepLabel := GetStepLabel(m.moduleName, m.stepNumber)
		numberedStepLabel := fmt.Sprintf("%d. %s", m.stepNumber+1, currentStepLabel)
		stepDetails = fmt.Sprintf("%s %s", numberedStepLabel, counters)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		currentlyPrefix,
		stepDetails,
	)

	return lipglossStyle.currentStepContainer.Render(content)
}

func getStepCounters(m Model) string {
	tm := GetTextManager()
	if tm == nil {
		// Fallback to hardcoded text if text manager not available
		return getStepCountersFallback(m)
	}

	moduleKey := strings.ToLower(m.moduleName)
	
	switch m.stepNumber {
	case 0:
		vars := i18n.CreateTemplateVars().
			SetString("count", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.chunkCount))).
			SetString("persons", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.personsFound))).
			SetString("mb", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.bytesReadMB)))
		return tm.GetCounterText(moduleKey, "creating_chunks", vars)

	case 1:
		var statusParts []string
		if m.firstChunkFixed {
			statusParts = append(statusParts, tm.GetStatusText("first_chunk"))
		}
		if m.lastChunkFixed {
			statusParts = append(statusParts, tm.GetStatusText("last_chunk"))
		}
		if len(statusParts) > 0 {
			vars := i18n.CreateTemplateVars().
				SetString("status", strings.Join(statusParts, ", "))
			return tm.GetCounterText(moduleKey, "fixing_structure_with_status", vars)
		}
		return tm.GetCounterText(moduleKey, "fixing_structure_no_status", nil)

	case 2:
		vars := i18n.CreateTemplateVars().
			SetString("persons", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.personsFound))).
			SetString("current", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.chunkCounter.Current))).
			SetString("total", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.chunkCount)))
		return tm.GetCounterText(moduleKey, "extracting_locations", vars)

	case 3:
		vars := i18n.CreateTemplateVars().
			SetString("coordinates", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.coordinateCount)))
		return tm.GetCounterText(moduleKey, "create_area_zones", vars)

	case 6:
		vars := i18n.CreateTemplateVars().
			SetString("coordinates", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.coordinateCount)))
		return tm.GetCounterText(moduleKey, "coordinates_only", vars)

	case 7:
		vars := i18n.CreateTemplateVars().
			SetString("current", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.binCounter.Current))).
			SetString("total", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.binCounter.Total)))
		return tm.GetCounterText(moduleKey, "bins_progress", vars)

	case 9:
		vars := i18n.CreateTemplateVars().
			SetString("agents", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.agentBinCount)))
		return tm.GetCounterText(moduleKey, "agents_count", vars)

	case 13:
		vars := i18n.CreateTemplateVars().
			SetString("current", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.scaleCounter.Current))).
			SetString("total", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.scaleCounter.Total)))
		return tm.GetCounterText(moduleKey, "scales_progress", vars)

	case 15:
		vars := i18n.CreateTemplateVars().
			SetString("scale", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.outputScale)))
		return tm.GetCounterText(moduleKey, "scale_value", vars)

	case 16:
		vars := i18n.CreateTemplateVars().
			SetString("current", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.outputCounter.Current))).
			SetString("total", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.outputCounter.Total)))
		return tm.GetCounterText(moduleKey, "files_progress", vars)

	case 18:
		if m.flagClean {
			vars := i18n.CreateTemplateVars().
				SetString("files", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.cleanupFiles))).
				SetString("dirs", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.cleanupDirs))).
				SetString("mb", lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.cleanupBytes)))
			return tm.GetCounterText(moduleKey, "cleanup_stats", vars)
		}
	}

	return "" // Default empty string if no counters for this step
}

// Fallback function for when text manager is not available
func getStepCountersFallback(m Model) string {
	switch m.stepNumber {
	case 0:
		return fmt.Sprintf("1. Creating %s smaller chunks: [%s persons found] [%s MB read]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.chunkCount)),
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.personsFound)),
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.bytesReadMB)))
	case 1:
		var statusParts []string
		if m.firstChunkFixed {
			statusParts = append(statusParts, "first chunk")
		}
		if m.lastChunkFixed {
			statusParts = append(statusParts, "last chunk")
		}
		if len(statusParts) > 0 {
			return fmt.Sprintf("2. Fixing XML structure: [%s fixed]", strings.Join(statusParts, ", "))
		}
		return "2. Fixing XML structure"
	case 2:
		return fmt.Sprintf("3. Extracting location from %s persons: [%s/%s]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.personsFound)),
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.chunkCounter.Current)),
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.chunkCount)))
	}
	return ""
}

func createSystemInfoSection(m Model) string {
	statusIndicator := "●"
	if m.statusBlink {
		statusIndicator = lipglossStyle.brightIndicator.Render("●")
	} else {
		statusIndicator = lipglossStyle.dimIndicator.Render("●")
	}

	tm := GetTextManager()
	var cpuText, ramText, systemLabel string
	
	if tm != nil {
		cpuValue := fmt.Sprintf("%.1f%%", m.cpuUsage)
		cpuVars := i18n.CreateTemplateVars().SetString("value", cpuValue)
		cpuText = lipglossStyle.cpuInfo.Render(tm.GetSystemText("cpu_format", cpuVars))

		// Converting RAM from MB to GB for better readability
		ramValueGB := m.ramUsage / 1024.0
		ramValue := fmt.Sprintf("%.2f GB", ramValueGB)
		ramVars := i18n.CreateTemplateVars().SetString("value", ramValue)
		ramText = lipglossStyle.ramInfo.Render(tm.GetSystemText("ram_format", ramVars))

		systemLabel = lipglossStyle.systemLabel.Render(tm.GetSystemText("label", nil))
	} else {
		// Fallback
		cpuValue := fmt.Sprintf("%.1f%%", m.cpuUsage)
		cpuText = lipglossStyle.cpuInfo.Render(fmt.Sprintf("[CPU: %s]", cpuValue))

		ramValueGB := m.ramUsage / 1024.0
		ramValue := fmt.Sprintf("%.2f GB", ramValueGB)
		ramText = lipglossStyle.ramInfo.Render(fmt.Sprintf("[RAM: %s]", ramValue))

		systemLabel = lipglossStyle.systemLabel.Render("SYSTEM: ")
	}

	systemInfo := systemLabel + cpuText + ramText
	content := fmt.Sprintf("%s%s", statusIndicator, systemInfo)

	return lipglossStyle.footerContainer.Render(content)
}

func createUpcomingStepsSection(m Model) string {
	var lines []string

	// Step count hardcoded to 18 based on current workflow implementation
	maxStep := 18
	for i := m.stepNumber + 1; i <= maxStep; i++ {
		if i == 3 && !m.flagDB {
			continue
		}

		if i == 18 && !m.flagClean {
			continue
		}

		stepLabel := GetStepLabel(m.moduleName, i)
		upcomingText := fmt.Sprintf("%d. %s", i+1, stepLabel)
		upcomingTask := lipglossStyle.upcomingTask.Render(upcomingText)
		lines = append(lines, upcomingTask)
	}

	if len(lines) == 0 {
		return "" // No upcoming steps
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
