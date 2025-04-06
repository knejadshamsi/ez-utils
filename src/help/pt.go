package help

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ptHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00A5FF")).
			Background(lipgloss.Color("#0A2A3F")).
			Padding(0, 1).
			Underline(true)

	ptFlagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#0A2A3F")).
			Padding(0, 1)

	ptTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			PaddingLeft(2)

	ptSectionStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(1).
			MarginBottom(1).
			Width(80)
)

func PrintPTHelp() {
	sections := []string{
		makePTUsageSection(),
		makePTDescriptionSection(),
		makePTEnvVarsSection(),
		makePTTimePeriodsSection(),
		makePTExamplesSection(),
	}
	fmt.Println(strings.Join(sections, "\n"))
}

func makePTUsageSection() string {
	usageWords := []string{"ez-utils", "pt", "<time-period>", "<gtfs-directory>"}
	styledUsageWords := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#C51010")).Render(usageWords[0]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render(usageWords[1]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCC00")).Render(usageWords[2]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCC00")).Render(usageWords[3]),
	}

	return ptSectionStyle.Render(
		ptHeaderStyle.Render("USAGE") + "\n\n" +
			strings.Join(styledUsageWords, " "),
	)
}

func makePTDescriptionSection() string {
	return ptSectionStyle.Render(
		ptHeaderStyle.Render("DESCRIPTION") + "\n\n" +
			ptTextStyle.Render("Sophisticated GTFS data processor that performs multi-step transit analysis through a sequential pipeline. Handles file validation, service pattern analysis, route mapping, and stop sequence processing. Features memory-efficient streaming operations, robust validation, and intelligent service selection based on the specified time period."),
	)
}

func makePTEnvVarsSection() string {
	envVars := []string{
		ptFlagStyle.Render("DB_HOST") + ": Database host address",
		ptFlagStyle.Render("DB_PORT") + ": Database port",
		ptFlagStyle.Render("DB_NAME") + ": Database name",
		ptFlagStyle.Render("DB_USER") + ": Database username",
		ptFlagStyle.Render("DB_PASSWORD") + ": Database password",
	}

	return ptSectionStyle.Render(
		ptHeaderStyle.Render("ENVIRONMENT VARIABLES") + "\n\n" +
			ptTextStyle.Render(strings.Join(envVars, "\n")),
	)
}

func makePTTimePeriodsSection() string {
	periods := []string{
		ptFlagStyle.Render("week") + ": Selects and processes ONE random day from Saturday or Sunday",
		ptFlagStyle.Render("work") + ": Selects and processes ONE random day from Monday through Friday",
		ptFlagStyle.Render("DD-MM-YY") + ": Processes data for a specific date (e.g., 15-04-25)",
	}

	return ptSectionStyle.Render(
		ptHeaderStyle.Render("TIME PERIODS") + "\n\n" +
			ptTextStyle.Render(strings.Join(periods, "\n")),
	)
}

func makePTExamplesSection() string {
	examples := []string{
		"ez-utils pt week ./gtfs-data",
		"ez-utils pt work ./gtfs-data",
		"ez-utils pt 15-04-25 ./gtfs-data",
	}

	return ptSectionStyle.Render(
		ptHeaderStyle.Render("EXAMPLES") + "\n\n" +
			ptTextStyle.Render(strings.Join(examples, "\n")),
	)
}
