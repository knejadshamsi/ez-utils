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
		makePTConfigSection(),
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

func makePTConfigSection() string {
	configInfo := []string{
		"Configuration is managed through a YAML file. Priority order:",
		"  1. Local: ./config.yaml (project-specific)",
		"  2. Global: Standard OS location (shared across projects)",
		"",
		"Global config locations:",
		"  • Linux/Unix: ~/.config/ez-utils/config.yaml",
		"  • Windows: %APPDATA%/ez-utils/config.yaml",
		"  • macOS: ~/Library/Application Support/ez-utils/config.yaml",
		"",
		"Use " + ptFlagStyle.Render("--new-config") + " to create a new config file interactively",
		"",
		"Database settings (required for PT analysis):",
		ptFlagStyle.Render("database.host") + ": Database host address",
		ptFlagStyle.Render("database.port") + ": Database port (default: 5432)",
		ptFlagStyle.Render("database.name") + ": Database name",
		ptFlagStyle.Render("database.user") + ": Database username",
		ptFlagStyle.Render("database.password") + ": Database password",
	}

	return ptSectionStyle.Render(
		ptHeaderStyle.Render("CONFIGURATION") + "\n\n" +
			ptTextStyle.Render(strings.Join(configInfo, "\n")),
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
