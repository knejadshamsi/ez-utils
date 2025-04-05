package help

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func PrintPTHelp() {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00A5FF")).
		Underline(true)

	flagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF"))

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#222222")).
		PaddingLeft(2).
		PaddingRight(2)

	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#888888")).
		Padding(1)

	usage := sectionStyle.Render(
		headerStyle.Render("USAGE") + "\n\n" +
			textStyle.Render("ez-utils pt <time-period> <gtfs-directory>"))

	description := sectionStyle.Render(
		headerStyle.Render("DESCRIPTION") + "\n\n" +
			textStyle.Render("Sophisticated GTFS data processor that performs multi-step transit analysis through a sequential pipeline. Handles file validation, service pattern analysis, route mapping, and stop sequence processing. Features memory-efficient streaming operations, robust validation, and intelligent service selection based on the specified time period."))

	envVars := sectionStyle.Render(
		headerStyle.Render("ENVIRONMENT VARIABLES") + "\n\n" +
			flagStyle.Render("DB_HOST") + ": Database host address\n" +
			flagStyle.Render("DB_PORT") + ": Database port\n" +
			flagStyle.Render("DB_NAME") + ": Database name\n" +
			flagStyle.Render("DB_USER") + ": Database username\n" +
			flagStyle.Render("DB_PASSWORD") + ": Database password")

	timePeriods := sectionStyle.Render(
		headerStyle.Render("TIME PERIODS") + "\n\n" +
			flagStyle.Render("week") + ": Selects and processes ONE random day from Saturday or Sunday\n" +
			flagStyle.Render("work") + ": Selects and processes ONE random day from Monday through Friday\n" +
			flagStyle.Render("DD-MM-YY") + ": Processes data for a specific date (e.g., 15-04-25)")

	examples := sectionStyle.Render(
		headerStyle.Render("EXAMPLES") + "\n\n" +
			textStyle.Render("ez-utils pt week ./gtfs-data\n"+
				"ez-utils pt work ./gtfs-data\n"+
				"ez-utils pt 15-04-25 ./gtfs-data"))

	output := lipgloss.JoinVertical(lipgloss.Left,
		usage,
		description,
		envVars,
		timePeriods,
		examples,
	)

	fmt.Println(output)
}
