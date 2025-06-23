package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"ez-utils/src/display/lang"
	displaylang "ez-utils/src/display" // Alias for top-level lang.go
)

// Handles the UI rendering
func (m Model) View() string {
	if m.ProcessComplete { // Changed to m.ProcessComplete
		// Process complete - return empty string to hide display
		return ""
	}

	titleSection := CreateTitleSection(m)
	previousStepsSection := CreatePreviousStepsSection(m)
	currentStepSection := CreateCurrentStepSection(m)
	systemInfoSection := CreateSystemInfoSection(m)
	upcomingStepsSection := CreateUpcomingStepsSection(m)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleSection,
		previousStepsSection,
		currentStepSection,
		systemInfoSection,
		upcomingStepsSection,
	)
}

func CreateTitleSection(m Model) string { // Made public
	tm := displaylang.GetTextManager() // Changed to displaylang.GetTextManager()
	if tm == nil {
		// Use configured process title instead of hardcoded text
		ezPart := "[EZ-UTILS]"
		descPart := ": " + m.ProcessTitle + " " // Changed to m.ProcessTitle
		timeStr := fmt.Sprintf("[ELAPSED TIME %02d:%02d]", int(m.TotalElapsedTime.Minutes()), int(m.TotalElapsedTime.Seconds())%60) // Changed to m.TotalElapsedTime
		
		titleContent := LipglossStyle.TitleEZ.Render(ezPart) + LipglossStyle.TitleText.Render(descPart) // Changed to LipglossStyle
		rightContent := LipglossStyle.TimeInfo.Render(timeStr) // Changed to LipglossStyle
		return titleContent + rightContent
	}

	ezPart := tm.GetUIText("title_prefix")
	// Use configured process title instead of locale hardcoded text
	descPart := ": " + m.ProcessTitle + " " // Changed to m.ProcessTitle
	timeStr := tm.FormatTimeText(int(m.TotalElapsedTime.Minutes()), int(m.TotalElapsedTime.Seconds())%60) // Changed to m.TotalElapsedTime

	titleContent := LipglossStyle.TitleEZ.Render(ezPart) + LipglossStyle.TitleText.Render(descPart) // Changed to LipglossStyle
	rightContent := LipglossStyle.TimeInfo.Render(timeStr) // Changed to LipglossStyle

	return titleContent + rightContent
}

func CreatePreviousStepsSection(m Model) string { // Made public
	var lines []string

	for i := 0; i < m.StepNumber; i++ { // Changed to m.StepNumber
		stepLabel := displaylang.GetStepLabelFromConfig(m.Steps, m.ModuleName, i) // Changed to displaylang.GetStepLabelFromConfig, m.Steps, m.ModuleName

		completedText := fmt.Sprintf("%d. %s", i+1, stepLabel)

		var durationStr string
		if duration, ok := m.StepDurations[i]; ok { // Changed to m.StepDurations
			durationStr = formatDuration(duration)
		} else {
			durationStr = "less than a second" // Default if no duration stored
		}

		completedTaskText := LipglossStyle.CompletedTask.Render(completedText) // Changed to LipglossStyle
		durationText := LipglossStyle.CompletedDuration.Render(fmt.Sprintf("{%s}", durationStr)) // Changed to LipglossStyle

		completedLine := fmt.Sprintf("%s %s", completedTaskText, durationText)
		lines = append(lines, completedLine)
	}

	if len(lines) == 0 {
		return "" // No completed steps yet
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func CreateCurrentStepSection(m Model) string { // Made public
	tm := displaylang.GetTextManager() // Changed to displaylang.GetTextManager()
	var currentlyText string
	if tm != nil {
		currentlyText = tm.GetUIText("currently_prefix")
	} else {
		currentlyText = "ACTION"
	}
	
	actionPrefix := LipglossStyle.CurrentlyPrefix.Render(currentlyText) // Changed to LipglossStyle
	spinnerView := LipglossStyle.Spinner.Render(m.Spinner.View()) // Changed to LipglossStyle, m.Spinner
	
	counters := GetStepCounters(m) // Changed to GetStepCounters

	var stepLabel string
	// Special case: First 4 steps have counter text that includes step number and description
	if m.StepNumber == 0 || m.StepNumber == 1 || m.StepNumber == 2 || m.StepNumber == 3 { // Changed to m.StepNumber
		stepLabel = counters
	} else {
		// For other steps, add the numbered step label plus the counters
		currentStepLabel := displaylang.GetStepLabelFromConfig(m.Steps, m.ModuleName, m.StepNumber) // Changed to displaylang.GetStepLabelFromConfig, m.Steps, m.ModuleName, m.StepNumber
		numberedStepLabel := fmt.Sprintf("%d. %s", m.StepNumber+1, currentStepLabel) // Changed to m.StepNumber
		stepLabel = fmt.Sprintf("%s %s", numberedStepLabel, counters)
	}

	// Combine all elements in a single line: ACTION [spinner] Step Title + counters
	content := fmt.Sprintf("%s %s %s", actionPrefix, spinnerView, stepLabel)

	return LipglossStyle.CurrentStepContainer.Render(content) // Changed to LipglossStyle
}

func GetStepCounters(m Model) string { // Made public
	tm := displaylang.GetTextManager() // Changed to displaylang.GetTextManager()
	if tm == nil {
		// Fallback to hardcoded text if text manager not available
		return GetStepCountersFallback(m) // Changed to GetStepCountersFallback
	}

	moduleKey := strings.ToLower(m.ModuleName) // Changed to m.ModuleName
	
	switch m.StepNumber { // Changed to m.StepNumber
	case 0:
		vars := lang.CreateTemplateVars().
			SetString("count", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.ChunkCount))). // Changed to LipglossStyle, m.ChunkCount
			SetString("persons", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.PersonsFound))). // Changed to LipglossStyle, m.PersonsFound
			SetString("mb", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.BytesReadMB))) // Changed to LipglossStyle, m.BytesReadMB
		return tm.GetCounterText(moduleKey, "creating_chunks", vars)

	case 1:
		var statusParts []string
		if m.FirstChunkFixed { // Changed to m.FirstChunkFixed
			statusParts = append(statusParts, tm.GetStatusText("first_chunk"))
		}
		if m.LastChunkFixed { // Changed to m.LastChunkFixed
			statusParts = append(statusParts, tm.GetStatusText("last_chunk"))
		}
		if len(statusParts) > 0 {
			vars := lang.CreateTemplateVars().
				SetString("status", strings.Join(statusParts, ", "))
			return tm.GetCounterText(moduleKey, "fixing_structure_with_status", vars)
		}
		return tm.GetCounterText(moduleKey, "fixing_structure_no_status", nil)

	case 2:
		vars := lang.CreateTemplateVars().
			SetString("persons", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.PersonsFound))). // Changed to LipglossStyle, m.PersonsFound
			SetString("current", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.ChunkCounter.Current))). // Changed to LipglossStyle, m.ChunkCounter.Current
			SetString("total", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.ChunkCount))) // Changed to LipglossStyle, m.ChunkCount
		return tm.GetCounterText(moduleKey, "extracting_locations", vars)

	case 3:
		vars := lang.CreateTemplateVars().
			SetString("coordinates", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.CoordinateCount))) // Changed to LipglossStyle, m.CoordinateCount
		return tm.GetCounterText(moduleKey, "create_area_zones", vars)

	case 6:
		vars := lang.CreateTemplateVars().
			SetString("coordinates", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.CoordinateCount))) // Changed to LipglossStyle, m.CoordinateCount
		return tm.GetCounterText(moduleKey, "coordinates_only", vars)

	case 7:
		vars := lang.CreateTemplateVars().
			SetString("current", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.BinCounter.Current))). // Changed to LipglossStyle, m.BinCounter.Current
			SetString("total", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.BinCounter.Total))) // Changed to LipglossStyle, m.BinCounter.Total
		return tm.GetCounterText(moduleKey, "bins_progress", vars)

	case 9:
		vars := lang.CreateTemplateVars().
			SetString("agents", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.AgentBinCount))) // Changed to LipglossStyle, m.AgentBinCount
		return tm.GetCounterText(moduleKey, "agents_count", vars)

	case 13:
		vars := lang.CreateTemplateVars().
			SetString("current", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.ScaleCounter.Current))). // Changed to LipglossStyle, m.ScaleCounter.Current
			SetString("total", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.ScaleCounter.Total))) // Changed to LipglossStyle, m.ScaleCounter.Total
		return tm.GetCounterText(moduleKey, "scales_progress", vars)

	case 15:
		vars := lang.CreateTemplateVars().
			SetString("scale", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.OutputScale))) // Changed to LipglossStyle, m.OutputScale
		return tm.GetCounterText(moduleKey, "scale_value", vars)

	case 16:
		vars := lang.CreateTemplateVars().
			SetString("current", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.OutputCounter.Current))). // Changed to LipglossStyle, m.OutputCounter.Current
			SetString("total", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.OutputCounter.Total))) // Changed to LipglossStyle, m.OutputCounter.Total
		return tm.GetCounterText(moduleKey, "files_progress", vars)

	case 18:
		if m.GetFlag("clean") {
			vars := lang.CreateTemplateVars().
				SetString("files", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.CleanupFiles))).
				SetString("dirs", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.CleanupDirs))).
				SetString("mb", LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.CleanupBytes)))
			return tm.GetCounterText(moduleKey, "cleanup_stats", vars)
		}
	}

	return "" // Default empty string if no counters for this step
}

// Fallback function for when text manager is not available
func GetStepCountersFallback(m Model) string {
	switch m.StepNumber {
	case 0:
		return fmt.Sprintf("1. Creating %s smaller chunks: [%s persons found] [%s MB read]",
			LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.ChunkCount)),
			LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.PersonsFound)),
			LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.BytesReadMB)))
	case 1:
		var statusParts []string
		if m.FirstChunkFixed {
			statusParts = append(statusParts, "first chunk")
		}
		if m.LastChunkFixed {
			statusParts = append(statusParts, "last chunk")
		}
		if len(statusParts) > 0 {
			return fmt.Sprintf("2. Fixing XML structure: [%s fixed]", strings.Join(statusParts, ", "))
		}
		return "2. Fixing XML structure"
	case 2:
		return fmt.Sprintf("3. Extracting location from %s persons: [%s/%s]",
			LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.PersonsFound)),
			LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.ChunkCounter.Current)),
			LipglossStyle.CounterValue.Render(fmt.Sprintf("%d", m.ChunkCounter.Total)))
	}
	return ""
}

func CreateSystemInfoSection(m Model) string { // Made public
	statusIndicator := "●"
	if m.StatusBlink { // Changed to m.StatusBlink
		statusIndicator = LipglossStyle.BrightIndicator.Render("●") // Changed to LipglossStyle
	} else {
		statusIndicator = LipglossStyle.DimIndicator.Render("●") // Changed to LipglossStyle
	}

	tm := displaylang.GetTextManager() // Changed to displaylang.GetTextManager()
	var systemLabel, abortText, leftContent string

	if tm != nil {
		var systemInfoParts []string
		systemLabel = LipglossStyle.SystemLabel.Render(tm.GetSystemText("label", nil))
		systemInfoParts = append(systemInfoParts, systemLabel)

		if m.CPUUsage != 0.0 {
			cpuValue := fmt.Sprintf("%.1f%%", m.CPUUsage)
			cpuVars := lang.CreateTemplateVars().SetString("value", cpuValue)
			systemInfoParts = append(systemInfoParts, LipglossStyle.CPUInfo.Render(tm.GetSystemText("cpu_format", cpuVars)))
		}

		if m.RAMUsage != 0.0 {
			ramValueGB := m.RAMUsage / 1024.0
			ramValue := fmt.Sprintf("%.2f GB", ramValueGB)
			ramVars := lang.CreateTemplateVars().SetString("value", ramValue)
			systemInfoParts = append(systemInfoParts, LipglossStyle.RAMInfo.Render(tm.GetSystemText("ram_format", ramVars)))
		}

		abortText = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(tm.GetUIText("abort_hint"))
		systemInfo := strings.Join(systemInfoParts, "")
		leftContent = fmt.Sprintf("%s%s", statusIndicator, systemInfo)
	} else {
		// Fallback
		var systemInfoParts []string
		systemLabel = LipglossStyle.SystemLabel.Render("SYSTEM: ")
		systemInfoParts = append(systemInfoParts, systemLabel)

		if m.CPUUsage != 0.0 {
			cpuValue := fmt.Sprintf("%.1f%%", m.CPUUsage)
			systemInfoParts = append(systemInfoParts, LipglossStyle.CPUInfo.Render(fmt.Sprintf("[CPU: %s]", cpuValue)))
		}

		if m.RAMUsage != 0.0 {
			ramValueGB := m.RAMUsage / 1024.0
			ramValue := fmt.Sprintf("%.2f GB", ramValueGB)
			systemInfoParts = append(systemInfoParts, LipglossStyle.RAMInfo.Render(fmt.Sprintf("[RAM: %s]", ramValue)))
		}

		abortText = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render("[Press q to abort]")
		systemInfo := strings.Join(systemInfoParts, "")
		leftContent = fmt.Sprintf("%s%s", statusIndicator, systemInfo)
	}
	
	// Use lipgloss to create a justified layout
	content := lipgloss.JoinHorizontal(lipgloss.Left, leftContent, " ", abortText)

	return LipglossStyle.FooterContainer.Render(content)
}

func CreateUpcomingStepsSection(m Model) string { // Made public
	var lines []string

	// Use configurable max steps
	maxStep := m.MaxSteps // Changed to m.MaxSteps
	if maxStep <= 0 {
		// Invalid configuration, skip upcoming steps
		return ""
	}

	for i := m.StepNumber + 1; i < maxStep; i++ { // Changed to m.StepNumber
		// Check flags dynamically using the new flags map
		if i == 3 && !m.GetFlag("db") { // Changed to m.GetFlag
			continue
		}

		if i == 18 && !m.GetFlag("clean") { // Changed to m.GetFlag
			continue
		}

		stepLabel := displaylang.GetStepLabelFromConfig(m.Steps, m.ModuleName, i) // Changed to displaylang.GetStepLabelFromConfig, m.Steps, m.ModuleName
		upcomingText := fmt.Sprintf("%d. %s", i+1, stepLabel)
		upcomingTask := LipglossStyle.UpcomingTask.Render(upcomingText) // Changed to LipglossStyle
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
func getModelFlag(m Model, flag string) bool { // This function is now a method on the Model struct
	return m.GetFlag(flag)
}
