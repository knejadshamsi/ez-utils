// Step 2: Service Pattern Identification
// This file focuses on identifying service patterns within the GTFS calendar data.
// It reads the calendar.txt file to determine which service IDs operate on the specified service day.
package pt

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (pto *PTOrchestrator) identifyServicePatterns() error {
	pto.logMessage("Starting Service Pattern Identification")

	// Read calendar.txt file
	calendarPath := filepath.Join(pto.config.GTFSDirectory, CalendarFile)
	file, err := os.Open(calendarPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open %s: %v", CalendarFile, err))
		return NewFileParsingError(CalendarFile, err.Error())
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read %s headers: %v", CalendarFile, err))
		return NewFileParsingError(CalendarFile, "invalid header row")
	}

	// Create header index map
	headerIndex := make(map[string]int)
	for i, header := range headers {
		// Strip UTF-8 BOM from header if present (fixes STM GTFS data)
		cleanHeader := strings.TrimPrefix(header, "\uFEFF")
		headerIndex[cleanHeader] = i
	}

	// Verify required columns exist
	requiredColumns := []string{"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday", "start_date", "end_date"}
	for _, col := range requiredColumns {
		if _, exists := headerIndex[col]; !exists {
			pto.logMessage(fmt.Sprintf("Missing required column in %s: %s", CalendarFile, col))
			return NewFileParsingError(CalendarFile, fmt.Sprintf("missing column: %s", col))
		}
	}

	// Step 1: If no target date, find smart weekday selection
	targetDate := pto.config.TargetDate
	if targetDate == "" {
		selectedDate, err := pto.selectSmartWeekday(reader, headerIndex)
		if err != nil {
			return err
		}
		targetDate = selectedDate
		pto.config.TargetDate = targetDate
		pto.logMessage(fmt.Sprintf("Auto-selected target date: %s", targetDate))
	} else {
		pto.logMessage(fmt.Sprintf("Using provided target date: %s", targetDate))
	}

	// Step 2: Process services for the selected single day
	file.Seek(0, 0) // Reset to beginning
	reader = csv.NewReader(file)
	reader.Read() // Skip header

	var servicePatterns []ServicePattern
	serviceCounter := 0

	// Read data rows and filter for single day
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pto.logMessage(fmt.Sprintf("Error reading %s row: %v", CalendarFile, err))
			continue
		}

		// Parse service pattern with date fields
		pattern := ServicePattern{
			ServiceID: record[headerIndex["service_id"]],
			Monday:    record[headerIndex["monday"]] == "1",
			Tuesday:   record[headerIndex["tuesday"]] == "1",
			Wednesday: record[headerIndex["wednesday"]] == "1",
			Thursday:  record[headerIndex["thursday"]] == "1",
			Friday:    record[headerIndex["friday"]] == "1",
			Saturday:  record[headerIndex["saturday"]] == "1",
			Sunday:    record[headerIndex["sunday"]] == "1",
			StartDate: record[headerIndex["start_date"]],
			EndDate:   record[headerIndex["end_date"]],
		}

		// Filter: Only services operating on target date
		if pto.operatesOnDate(pattern, targetDate) {
			servicePatterns = append(servicePatterns, pattern)
			serviceCounter++

			pto.stats.ServiceCounter = serviceCounter

			if serviceCounter%5 == 0 {
				pto.logMessage(fmt.Sprintf("Found %d services for %s (last: %s)", serviceCounter, targetDate, pattern.ServiceID))
			}
		}

		if pto.tuiEnabled && pto.tui != nil {
			pto.tui.UpdateCounter("services", fmt.Sprintf("%d", serviceCounter))
		}
	}

	if len(servicePatterns) == 0 {
		pto.logMessage("No matching service patterns found")
		return NewInvalidServiceError(pto.getServiceDayString())
	}

	// Create output directory and file
	outputDir := filepath.Join(pto.config.TempDir, ServicesDir)
	if err := CreateDirectoryIfNotExists(outputDir); err != nil {
		return NewDirectoryError(outputDir, "create")
	}
	outputPath := filepath.Join(outputDir, ServicePatternsFile)
	jsonData, err := json.MarshalIndent(servicePatterns, "", "  ")
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to marshal service patterns to JSON: %v", err))
		return NewFileParsingError(ServicePatternsFile, err.Error())
	}

	if err := os.WriteFile(outputPath, jsonData, FilePermissions); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to write service patterns file: %v", err))
		return NewFileParsingError(ServicePatternsFile, err.Error())
	}

	// Final live update with completion status
	if pto.tuiEnabled && pto.tui != nil {
		pto.tui.UpdateCounter("services", fmt.Sprintf("%d", serviceCounter))
		pto.tui.UpdateCounter("progress", "100%")
	}

	pto.logMessage(fmt.Sprintf("Service pattern identification completed successfully. Found %d matching patterns", serviceCounter))
	return nil
}

// matchesServiceDay checks if a service pattern matches the configured service day
func (pto *PTOrchestrator) matchesServiceDay(pattern ServicePattern) bool {
	switch pto.config.ServiceDay {
	case ServiceDayWeekday:
		return pattern.Monday || pattern.Tuesday || pattern.Wednesday || pattern.Thursday || pattern.Friday
	case ServiceDayWeekend:
		return pattern.Saturday || pattern.Sunday
	case ServiceDayMonday:
		return pattern.Monday
	case ServiceDayTuesday:
		return pattern.Tuesday
	case ServiceDayWednesday:
		return pattern.Wednesday
	case ServiceDayThursday:
		return pattern.Thursday
	case ServiceDayFriday:
		return pattern.Friday
	case ServiceDaySaturday:
		return pattern.Saturday
	case ServiceDaySunday:
		return pattern.Sunday
	default:
		return false
	}
}

// getServiceDayString returns a string representation of the service day for error messages
func (pto *PTOrchestrator) getServiceDayString() string {
	switch pto.config.ServiceDay {
	case ServiceDayWeekday:
		return "weekday"
	case ServiceDayWeekend:
		return "weekend"
	case ServiceDayMonday:
		return "monday"
	case ServiceDayTuesday:
		return "tuesday"
	case ServiceDayWednesday:
		return "wednesday"
	case ServiceDayThursday:
		return "thursday"
	case ServiceDayFriday:
		return "friday"
	case ServiceDaySaturday:
		return "saturday"
	case ServiceDaySunday:
		return "sunday"
	default:
		return "unknown"
	}
}

// selectSmartWeekday picks a random weekday with both bus and metro service
func (pto *PTOrchestrator) selectSmartWeekday(reader *csv.Reader, headerIndex map[string]int) (string, error) {
	pto.logMessage("Auto-selecting weekday with bus and metro service...")

	// Read all services to find valid weekdays
	var allServices []ServicePattern
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		pattern := ServicePattern{
			ServiceID: record[headerIndex["service_id"]],
			Monday:    record[headerIndex["monday"]] == "1",
			Tuesday:   record[headerIndex["tuesday"]] == "1",
			Wednesday: record[headerIndex["wednesday"]] == "1",
			Thursday:  record[headerIndex["thursday"]] == "1",
			Friday:    record[headerIndex["friday"]] == "1",
			Saturday:  record[headerIndex["saturday"]] == "1",
			Sunday:    record[headerIndex["sunday"]] == "1",
			StartDate: record[headerIndex["start_date"]],
			EndDate:   record[headerIndex["end_date"]],
		}

		// Only consider weekday services
		if pattern.Monday || pattern.Tuesday || pattern.Wednesday || pattern.Thursday || pattern.Friday {
			allServices = append(allServices, pattern)
		}
	}

	if len(allServices) == 0 {
		return "", fmt.Errorf("no weekday services found")
	}

	// Find first service with both bus and metro, pick a valid date
	for _, service := range allServices {
		if pto.hasCompleteTransitCoverage(service.ServiceID) {
			// Pick a random weekday from this service's date range
			validDate := pto.pickRandomWeekdayInRange(service.StartDate, service.EndDate)
			if validDate != "" {
				return validDate, nil
			}
		}
	}

	// Fallback: use first service's date range
	service := allServices[0]
	validDate := pto.pickRandomWeekdayInRange(service.StartDate, service.EndDate)
	if validDate != "" {
		pto.logMessage("Warning: Selected service may not have complete bus+metro coverage")
		return validDate, nil
	}

	return "", fmt.Errorf("no valid weekday dates found")
}

// operatesOnDate checks if service operates on specific date
func (pto *PTOrchestrator) operatesOnDate(pattern ServicePattern, targetDate string) bool {
	// Convert YYYY-MM-DD to YYYYMMDD
	gtfsDate := strings.ReplaceAll(targetDate, "-", "")

	// Check date range
	if gtfsDate < pattern.StartDate || gtfsDate > pattern.EndDate {
		return false
	}

	// Parse target date to get day of week
	parsedDate, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return false
	}

	// Check if service operates on this day of week
	switch parsedDate.Weekday() {
	case time.Monday:
		return pattern.Monday
	case time.Tuesday:
		return pattern.Tuesday
	case time.Wednesday:
		return pattern.Wednesday
	case time.Thursday:
		return pattern.Thursday
	case time.Friday:
		return pattern.Friday
	case time.Saturday:
		return pattern.Saturday
	case time.Sunday:
		return pattern.Sunday
	}
	return false
}

// hasCompleteTransitCoverage checks if service has both bus and metro (simplified)
func (pto *PTOrchestrator) hasCompleteTransitCoverage(serviceID string) bool {
	// For now, assume first service found has complete coverage
	// In reality, would need to check service-route mappings
	return true
}

// pickRandomWeekdayInRange picks random weekday from GTFS date range
func (pto *PTOrchestrator) pickRandomWeekdayInRange(startDate, endDate string) string {
	// Parse GTFS dates (YYYYMMDD)
	start, err := time.Parse("20060102", startDate)
	if err != nil {
		return ""
	}
	end, err := time.Parse("20060102", endDate)
	if err != nil {
		return ""
	}

	// Find weekdays in range
	var weekdays []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if d.Weekday() >= time.Monday && d.Weekday() <= time.Friday {
			weekdays = append(weekdays, d)
		}
	}

	if len(weekdays) == 0 {
		return ""
	}

	// Pick random weekday
	rand.Seed(time.Now().UnixNano())
	selected := weekdays[rand.Intn(len(weekdays))]
	return selected.Format("2006-01-02")
}
