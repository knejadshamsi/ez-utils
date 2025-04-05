package help

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#38B0DE")).
			Underline(true)

	flagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#222222")).
			PaddingLeft(2).
			PaddingRight(2)

	sectionStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(1).
			MarginBottom(1)
)

func PrintPopulationHelp() {
	sections := []string{
		makeUsageSection(),
		makeDescriptionSection(),
		makeEnvVarsSection(),
		makeFlagsSection(),
		makeExamplesSection(),
	}

	fmt.Println(strings.Join(sections, "\n"))
}

func makeUsageSection() string {
	return sectionStyle.Render(
		headerStyle.Render("USAGE") + "\n\n" +
			descStyle.Render("ez-utils population <directory>"),
	)
}

func makeDescriptionSection() string {
	return sectionStyle.Render(
		headerStyle.Render("DESCRIPTION") + "\n\n" +
			descStyle.Render("High-performance XML population processor that handles chunking, coordinate transformation (MTM8 to WGS84), spatial binning, and agent scaling. Processes XML files from the input directory and outputs MATSim-compatible files in the same location."),
	)
}

func makeEnvVarsSection() string {
	envVars := []string{
		"CHUNK_SIZE: Defines the size of XML chunks for processing",
		"DB_HOST: Database host address",
		"DB_PORT: Database port",
		"DB_NAME: Database name",
		"DB_USER: Database username",
		"DB_PASSWORD: Database password",
	}

	return sectionStyle.Render(
		headerStyle.Render("ENVIRONMENT VARIABLES") + "\n\n" +
			descStyle.Render(strings.Join(envVars, "\n")),
	)
}

func makeFlagsSection() string {
	flags := []string{
		flagStyle.Render("-d, --db") + ": Enable PostgreSQL storage (optional)",
		flagStyle.Render("-s, --scale") + ": Target population scale factor (optional)\n  When not specified, automatically determines and produces up to 10 different scales (1% to 10%) based on population analysis",
		flagStyle.Render("-c, --clean") + ": Clean up temporary files after processing (optional)",
	}

	return sectionStyle.Render(
		headerStyle.Render("FLAGS") + "\n\n" +
			descStyle.Render(strings.Join(flags, "\n")),
	)
}

func makeExamplesSection() string {
	examples := []string{
		"ez-utils population ./my-population-data",
		"ez-utils population ./my-population-data -d --scale 0.1",
		"ez-utils population ./my-population-data --db -s 0.1 --clean",
	}

	return sectionStyle.Render(
		headerStyle.Render("EXAMPLES") + "\n\n" +
			descStyle.Render(strings.Join(examples, "\n")),
	)
}
