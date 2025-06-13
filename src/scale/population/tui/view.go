package tui

import (
	"fmt"
	"sort"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// View renders the complete TUI according to design specification
func (m UnifiedModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}
	
	var sections []string
	
	// Header Section
	sections = append(sections, m.renderHeader())
	
	// Active Workers Section
	sections = append(sections, m.renderActiveWorkers())
	
	// Controls
	sections = append(sections, m.renderControls())
	
	return strings.Join(sections, "\n")
}

// renderHeader renders the header section according to design specification
func (m UnifiedModel) renderHeader() string {
	var lines []string
	
	// Create dynamic title style based on safe terminal width
	safeWidth := m.safeWidth()
	dynamicTitleStyle := m.styles.titleStyle.Copy().Width(safeWidth)
	
	// Create border that fits the terminal width with safe calculation
	borderWidth := m.safeBorderWidth()
	topBorder := strings.Repeat("═", borderWidth)
	
	// 1. Top border
	lines = append(lines, dynamicTitleStyle.Render(topBorder))
	
	// 2. Title
	lines = append(lines, dynamicTitleStyle.Render("EZ-UTILS MATSim Population Scaling Pipeline"))
	
	// 3. Elapsed Time - centered, orange
	dynamicTimeStyle := m.styles.timeStyle.Copy().Width(safeWidth)
	lines = append(lines, dynamicTimeStyle.Render(fmt.Sprintf("[ELAPSED TIME: %s]", m.GetElapsedTime())))
	
	// 4. System Resources - colored individually
	metrics := m.GetSystemMetrics()
	cpuText := m.styles.cpuStyle.Render(fmt.Sprintf("[CPU %d%%]", metrics.CPU))
	ramText := m.styles.ramStyle.Render(fmt.Sprintf("[RAM %dMB]", metrics.RAM))
	diskText := m.styles.diskStyle.Render(fmt.Sprintf("[DISK %d%%]", metrics.DISK))
	systemLabel := m.styles.headerStyle.Copy().Bold(true).Render("SYSTEM:")
	systemLine := fmt.Sprintf("%s %s %s %s", systemLabel, cpuText, ramText, diskText)
	systemStyle := m.styles.headerStyle.Copy().Width(safeWidth).Align(lipgloss.Center)
	lines = append(lines, systemStyle.Render(systemLine))
	
	// 5. Bottom border (reuse same border string)
	lines = append(lines, dynamicTitleStyle.Render(topBorder))
	
	// 6. Phase - outside border, centered, white
	dynamicPhaseStyle := m.styles.phaseStyle.Copy().Width(safeWidth)
	lines = append(lines, dynamicPhaseStyle.Render(m.GetPhase()))
	
	// 7. Current Action - outside border, centered, yellow
	dynamicActionStyle := m.styles.actionStyle.Copy().Width(safeWidth)
	lines = append(lines, dynamicActionStyle.Render(fmt.Sprintf("[ACTION: %s]", m.currentAction)))
	
	return strings.Join(lines, "\n")
}

// renderActiveWorkers renders active workers according to design specification
func (m UnifiedModel) renderActiveWorkers() string {
	workers := m.GetActiveWorkers()
	
	if len(workers) == 0 {
		return ""
	}
	
	// Group workers by type
	workerGroups := make(map[WorkerType][]WorkerData)
	for _, worker := range workers {
		workerGroups[worker.WorkerType] = append(workerGroups[worker.WorkerType], worker)
	}
	
	var result []string
	
	// Get sorted worker types for deterministic ordering
	var workerTypes []WorkerType
	for workerType := range workerGroups {
		workerTypes = append(workerTypes, workerType)
	}
	sort.Slice(workerTypes, func(i, j int) bool {
		return int(workerTypes[i]) < int(workerTypes[j])
	})
	
	// Create tables for each worker type in deterministic order
	for _, workerType := range workerTypes {
		typeWorkers := workerGroups[workerType]
		if len(typeWorkers) == 0 {
			continue
		}
		
		// Sort workers within each type by name for consistent ordering
		sort.Slice(typeWorkers, func(i, j int) bool {
			return typeWorkers[i].Name < typeWorkers[j].Name
		})
		
		// Get headers and description for this worker type
		headers := getWorkerHeaders(workerType)
		description := getWorkerDescription(workerType)
		typeName := getWorkerTypeName(workerType)
		
		// Prepare table rows
		var rows [][]string
		for _, worker := range typeWorkers {
			rows = append(rows, []string{
				worker.Name,
				m.styles.GetStatusSymbol(worker.Status),
				worker.PrimaryMetric,
				worker.SecondaryMetric,
			})
		}
		
		// Create lipgloss table
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(colorWhite)).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					return lipgloss.NewStyle().Foreground(colorWhite).Bold(true)
				}
				return lipgloss.NewStyle().Foreground(colorWhite)
			}).
			Headers(headers...).
			Rows(rows...).
			Width(m.safeTableWidth()) // Account for padding with safe width
		
		result = append(result, "")
		result = append(result, m.styles.tableHeaderStyle.Render(fmt.Sprintf("%s Workers:", typeName)))
		
		// Add description for this worker type
		descStyle := m.styles.headerStyle.Copy().Width(m.safeTableWidth()).Italic(true)
		result = append(result, descStyle.Render(description))
		result = append(result, "")
		
		result = append(result, t.String())
	}
	
	return strings.Join(result, "\n")
}

// renderControls renders control instructions
func (m UnifiedModel) renderControls() string {
	// Only show quit control for production use
	return m.styles.headerStyle.Render("Press [q] to quit")
}

// WorkerTypeInfo contains display information for each worker type
type WorkerTypeInfo struct {
	Name        string
	Headers     []string
	Description string
}

// Worker type configuration map
var workerTypeConfig = map[WorkerType]WorkerTypeInfo{
	TypeReader: {
		Name:        "Reader",
		Headers:     []string{"Name", "Status", "Persons Found", "Lines Read"},
		Description: "Parse and read MATSim population XML files line by line. Extract raw person and activity data from the XML structure. Monitor file processing progress and data integrity during initial population loading.",
	},
	TypeExtractor: {
		Name:        "Extractor",
		Headers:     []string{"Name", "Status", "Persons Processed", "Persons Skipped"},
		Description: "Process parsed XML data to extract individual person records with their attributes and activities. Filter and validate person data quality, skipping malformed or incomplete records. Transform raw XML data into structured person objects for further processing.",
	},
	TypeMapper: {
		Name:        "Mapper",
		Headers:     []string{"Name", "Status", "Coords Processed", "Grid Bins"},
		Description: "Map person home coordinates to spatial grid bins for geographic distribution analysis. Create and maintain spatial indices for efficient location-based processing. Generate grid-based population density maps for scaling calculations.",
	},
	TypeReducer: {
		Name:        "Reducer",
		Headers:     []string{"Name", "Status", "Persons Processed", "Total Selected"},
		Description: "Apply scaling algorithms to select representative population subsets at target percentages. Process spatial and demographic criteria to maintain population representativeness. Generate final selection lists for each requested scale factor.",
	},
	TypeWriter: {
		Name:        "Writer",
		Headers:     []string{"Name", "Status", "Persons Written", "Scale Target"},
		Description: "Generate final scaled population XML files for each target percentage. Write properly formatted MATSim population files with selected persons and their complete activity chains. Validate output file structure and data completeness.",
	},
}

// Helper functions for worker type information
func getWorkerHeaders(workerType WorkerType) []string {
	if info, exists := workerTypeConfig[workerType]; exists {
		return info.Headers
	}
	return []string{"Name", "Status", "Primary Metric", "Secondary Metric"}
}

func getWorkerDescription(workerType WorkerType) string {
	if info, exists := workerTypeConfig[workerType]; exists {
		return info.Description
	}
	return "Generic worker performing data processing tasks."
}

func getWorkerTypeName(workerType WorkerType) string {
	if info, exists := workerTypeConfig[workerType]; exists {
		return info.Name
	}
	return "Unknown"
}