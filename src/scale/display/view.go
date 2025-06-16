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
		// Process complete - return empty string to hide display
		return ""
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
		// Use configured process title instead of hardcoded text
		ezPart := "[EZ-UTILS]"
		descPart := ": " + m.processTitle + " "
		timeStr := fmt.Sprintf("[ELAPSED TIME %02d:%02d]", int(m.totalElapsedTime.Minutes()), int(m.totalElapsedTime.Seconds())%60)
		
		titleContent := lipglossStyle.titleEZ.Render(ezPart) + lipglossStyle.titleText.Render(descPart)
		rightContent := lipglossStyle.timeInfo.Render(timeStr)
		return titleContent + rightContent
	}

	ezPart := tm.GetUIText("title_prefix")
	// Use configured process title instead of locale hardcoded text
	descPart := ": " + m.processTitle + " "
	timeStr := tm.FormatTimeText(int(m.totalElapsedTime.Minutes()), int(m.totalElapsedTime.Seconds())%60)

	titleContent := lipglossStyle.titleEZ.Render(ezPart) + lipglossStyle.titleText.Render(descPart)
	rightContent := lipglossStyle.timeInfo.Render(timeStr)

	return titleContent + rightContent
}

func createPreviousStepsSection(m Model) string {
	var lines []string

	for i := 0; i < m.stepNumber; i++ {
		stepLabel := GetStepLabelFromConfig(m.steps, m.moduleName, i)

		completedText := fmt.Sprintf("%d. %s", i+1, stepLabel)

		var durationStr string
		if duration, ok := m.stepDurations[i]; ok {
			durationStr = formatDuration(duration)
		} else {
			durationStr = "less than a second" // Default if no duration stored
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
		currentlyText = "ACTION"
	}
	
	actionPrefix := lipglossStyle.currentlyPrefix.Render(currentlyText)
	spinnerView := lipglossStyle.spinner.Render(m.spinner.View())
	
	counters := getStepCounters(m)

	var stepLabel string
	// Special case: First 4 steps have counter text that includes step number and description
	if m.stepNumber == 0 || m.stepNumber == 1 || m.stepNumber == 2 || m.stepNumber == 3 {
		stepLabel = counters
	} else {
		// For other steps, add the numbered step label plus the counters
		currentStepLabel := GetStepLabelFromConfig(m.steps, m.moduleName, m.stepNumber)
		numberedStepLabel := fmt.Sprintf("%d. %s", m.stepNumber+1, currentStepLabel)
		stepLabel = fmt.Sprintf("%s %s", numberedStepLabel, counters)
	}

	// Combine all elements in a single line: ACTION [spinner] Step Title + counters
	content := fmt.Sprintf("%s %s %s", actionPrefix, spinnerView, stepLabel)

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
		if getModelFlag(m, "clean") {
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
	var cpuText, ramText, diskText, systemLabel, abortText string
	
	if tm != nil {
		cpuValue := fmt.Sprintf("%.1f%%", m.cpuUsage)
		cpuVars := i18n.CreateTemplateVars().SetString("value", cpuValue)
		cpuText = lipglossStyle.cpuInfo.Render(tm.GetSystemText("cpu_format", cpuVars))

		// Converting RAM from MB to GB for better readability
		ramValueGB := m.ramUsage / 1024.0
		ramValue := fmt.Sprintf("%.2f GB", ramValueGB)
		ramVars := i18n.CreateTemplateVars().SetString("value", ramValue)
		ramText = lipglossStyle.ramInfo.Render(tm.GetSystemText("ram_format", ramVars))

		// Disk usage in GB
		diskValue := fmt.Sprintf("%.1f GB", m.diskUsage)
		diskVars := i18n.CreateTemplateVars().SetString("value", diskValue)
		diskText = lipglossStyle.diskInfo.Render(tm.GetSystemText("disk_format", diskVars))

		systemLabel = lipglossStyle.systemLabel.Render(tm.GetSystemText("label", nil))
		abortText = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(tm.GetUIText("abort_hint"))
	} else {
		// Fallback
		cpuValue := fmt.Sprintf("%.1f%%", m.cpuUsage)
		cpuText = lipglossStyle.cpuInfo.Render(fmt.Sprintf("[CPU: %s]", cpuValue))

		ramValueGB := m.ramUsage / 1024.0
		ramValue := fmt.Sprintf("%.2f GB", ramValueGB)
		ramText = lipglossStyle.ramInfo.Render(fmt.Sprintf("[RAM: %s]", ramValue))

		diskValue := fmt.Sprintf("%.1f GB", m.diskUsage)
		diskText = lipglossStyle.diskInfo.Render(fmt.Sprintf("[DISK: %s]", diskValue))

		systemLabel = lipglossStyle.systemLabel.Render("SYSTEM: ")
		abortText = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render("[Press q to abort]")
	}

	systemInfo := systemLabel + cpuText + ramText + diskText
	// Create a layout that places abort text on the right
	leftContent := fmt.Sprintf("%s%s", statusIndicator, systemInfo)
	
	// Use lipgloss to create a justified layout
	content := lipgloss.JoinHorizontal(lipgloss.Left, leftContent, " ", abortText)

	return lipglossStyle.footerContainer.Render(content)
}

func createUpcomingStepsSection(m Model) string {
	var lines []string

	// Use configurable max steps
	maxStep := m.maxSteps
	if maxStep <= 0 {
		// Invalid configuration, skip upcoming steps
		return ""
	}

	for i := m.stepNumber + 1; i < maxStep; i++ {
		// Check flags dynamically using the new flags map
		if i == 3 && !getModelFlag(m, "db") {
			continue
		}

		if i == 18 && !getModelFlag(m, "clean") {
			continue
		}

		stepLabel := GetStepLabelFromConfig(m.steps, m.moduleName, i)
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
	totalSeconds := int(d.Seconds())
	
	if totalSeconds < 1 {
		return "less than a second"
	} else if totalSeconds < 5 {
		return "few seconds"
	} else {
		minutes := int(d.Minutes())
		seconds := totalSeconds % 60
		return fmt.Sprintf("%02d:%02d", minutes, seconds)
	}
}

// getModelFlag safely retrieves a flag value from the model's flags map
func getModelFlag(m Model, flag string) bool {
	if m.flags == nil {
		return false
	}
	value, exists := m.flags[flag]
	return exists && value
}
