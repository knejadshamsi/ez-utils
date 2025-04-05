package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Handles the UI rendering
func (m Model) View() string {
	if m.processComplete {
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
	ezPart := "[EZ-UTILS]"
	descPart := ": Scaling down MATSim population "
	timeStr := fmt.Sprintf("[TIME %02d:%02d]", int(m.totalElapsedTime.Minutes()), int(m.totalElapsedTime.Seconds())%60)
	abortStr := " [Press q to abort]"

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
	currentlyPrefix := lipglossStyle.spinner.Render(m.spinner.View()) + " " + lipglossStyle.currentlyPrefix.Render("Currently")

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

	case 3:
		return fmt.Sprintf("4. Create Area Zones [%s coordinates]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.coordinateCount)))

	case 6:
		return fmt.Sprintf("[%s coordinates]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.coordinateCount)))

	case 7:
		return fmt.Sprintf("[%s/%s bins]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.binCounter.Current)),
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.binCounter.Total)))

	case 9:
		return fmt.Sprintf("[%s agents]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.agentBinCount)))

	case 13:
		return fmt.Sprintf("[%s/%s scales]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.scaleCounter.Current)),
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.scaleCounter.Total)))

	case 15:
		return fmt.Sprintf("[scale %s]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.outputScale)))

	case 16:
		return fmt.Sprintf("[%s/%s files]",
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.outputCounter.Current)),
			lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.outputCounter.Total)))

	case 18:
		if m.flagClean {
			return fmt.Sprintf("[%s files] [%s dirs] [%s MB]",
				lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.cleanupFiles)),
				lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.cleanupDirs)),
				lipglossStyle.counterValue.Render(fmt.Sprintf("%d", m.cleanupBytes)))
		}
	}

	return "" // Default empty string if no counters for this step
}

func createSystemInfoSection(m Model) string {
	statusIndicator := "●"
	if m.statusBlink {
		statusIndicator = lipglossStyle.brightIndicator.Render("●")
	} else {
		statusIndicator = lipglossStyle.dimIndicator.Render("●")
	}

	cpuValue := fmt.Sprintf("%.1f%%", m.cpuUsage)
	cpuText := lipglossStyle.cpuInfo.Render(fmt.Sprintf("[CPU: %s]", cpuValue))

	// Converting RAM from MB to GB for better readability
	ramValueGB := m.ramUsage / 1024.0
	ramValue := fmt.Sprintf("%.2f GB", ramValueGB)
	ramText := lipglossStyle.ramInfo.Render(fmt.Sprintf("[RAM: %s]", ramValue))

	systemInfo := lipglossStyle.systemLabel.Render("SYSTEM: ") + cpuText + ramText

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
