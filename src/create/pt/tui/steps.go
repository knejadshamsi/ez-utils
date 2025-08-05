package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RenderPrevious creates the completed steps section
func RenderPrevious(state *State) string {
	var lines []string

	for i := 0; i < state.CurrentStep; i++ {
		if i >= len(state.Steps) {
			continue
		}

		step := state.Steps[i]
		stepLabel := fmt.Sprintf("%d. %s", i+1, step.Name)

		var durationStr string
		if duration, ok := state.StepDurations[i]; ok {
			durationStr = formatDuration(duration)
		} else {
			durationStr = "less than a second"
		}

		completedTaskText := Styles.CompletedTask.Render(stepLabel)
		durationText := Styles.CompletedDuration.Render(fmt.Sprintf("{%s}", durationStr))

		completedLine := fmt.Sprintf("%s %s", completedTaskText, durationText)
		lines = append(lines, completedLine)
	}

	if len(lines) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// RenderOngoing creates the current step section with spinner and counters
func RenderOngoing(state *State) string {
	actionPrefix := Styles.CurrentlyPrefix.Render("ACTION")
	spinnerView := Styles.Spinner.Render(state.Spinner.View())

	counters := getStepCounters(state)

	var stepLabel string
	if state.CurrentStep < len(state.Steps) {
		currentStep := state.Steps[state.CurrentStep]
		numberedStepLabel := fmt.Sprintf("%d. %s", state.CurrentStep+1, currentStep.Name)
		if counters != "" {
			stepLabel = fmt.Sprintf("%s %s", numberedStepLabel, counters)
		} else {
			stepLabel = numberedStepLabel
		}
	} else {
		stepLabel = "Processing complete"
	}

	content := fmt.Sprintf("%s %s %s", actionPrefix, spinnerView, stepLabel)
	return Styles.CurrentStepContainer.Render(content)
}

// RenderUpcoming creates the upcoming steps section
func RenderUpcoming(state *State) string {
	var lines []string

	for i := state.CurrentStep + 1; i < state.TotalSteps && i < len(state.Steps); i++ {
		step := state.Steps[i]
		upcomingText := fmt.Sprintf("%d. %s", i+1, step.Name)
		upcomingTask := Styles.UpcomingTask.Render(upcomingText)
		lines = append(lines, upcomingTask)
	}

	if len(lines) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// getStepCounters returns the live counter text for the current step
func getStepCounters(state *State) string {
	if state.CurrentStep >= len(state.Steps) {
		return ""
	}

	counter := state.StepCounters[state.CurrentStep]

	switch state.CurrentStep {
	case 0: // Validate GTFS Files
		if counter.CurrentFile != "" {
			return fmt.Sprintf("[validating %s] [%s files] [%s KB]",
				Styles.CounterValue.Render(counter.CurrentFile),
				Styles.CounterValue.Render(fmt.Sprintf("%d", counter.FilesValidated)),
				Styles.CounterValue.Render(fmt.Sprintf("%d", counter.FileSize)))
		}
		return fmt.Sprintf("[%s files validated]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.FilesValidated)))

	case 1: // Identify Service Patterns
		if counter.Progress > 0 {
			return fmt.Sprintf("[%s services found] [%s%% progress]",
				Styles.CounterValue.Render(fmt.Sprintf("%d", counter.ServicesFound)),
				Styles.CounterValue.Render(fmt.Sprintf("%d", counter.Progress)))
		}
		return fmt.Sprintf("[%s services found]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.ServicesFound)))

	case 2: // Map Service Routes
		return fmt.Sprintf("[%s routes] [%s mappings]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.RoutesMapped)),
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.Mappings)))

	case 3: // Select Services
		return fmt.Sprintf("[%s/%s services selected]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.Selected)),
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.Total)))

	case 4: // Build Trip Database
		return fmt.Sprintf("[%s rows imported]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.RowsImported)))

	case 5: // Collect Trips
		if counter.TripProgress > 0 {
			return fmt.Sprintf("[%s trips] [%s%% progress]",
				Styles.CounterValue.Render(fmt.Sprintf("%d", counter.TripsCollected)),
				Styles.CounterValue.Render(fmt.Sprintf("%d", counter.TripProgress)))
		}
		return fmt.Sprintf("[%s trips collected]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.TripsCollected)))

	case 6: // Extract Stop Sequences
		return fmt.Sprintf("[%s sequences] [%s stops]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.Sequences)),
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.Stops)))

	case 7: // Gather Stop Details
		return fmt.Sprintf("[%s stop details]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.StopDetails)))

	case 8: // Generate XML
		if counter.XMLProgress > 0 {
			return fmt.Sprintf("[%s%% complete] [%s KB output]",
				Styles.CounterValue.Render(fmt.Sprintf("%d", counter.XMLProgress)),
				Styles.CounterValue.Render(fmt.Sprintf("%d", counter.OutputSize)))
		}
		return fmt.Sprintf("[%s KB output]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.OutputSize)))

	case 9: // Cleanup Files
		return fmt.Sprintf("[%s files] [%s dirs] [%s KB freed]",
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.FilesDeleted)),
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.DirsDeleted)),
			Styles.CounterValue.Render(fmt.Sprintf("%d", counter.BytesFreed/1024)))
	}

	return ""
}

// formatDuration formats a duration for display
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
