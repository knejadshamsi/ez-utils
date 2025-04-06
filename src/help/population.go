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
			Background(lipgloss.Color("#0A2A3F")).
			Padding(0, 1).
			Underline(true)

	flagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#0A2A3F")).
			Padding(0, 1)

	envVarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#0A2A3F")).
			Padding(0, 1)

	optionalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4040"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			PaddingLeft(2)

	sectionStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(1).
			MarginBottom(1).
			Width(80)

	combinedSectionStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				Padding(1).
				MarginBottom(1).
				Width(80)
)

func PrintPopulationHelp() {
	sections := []string{
		makeCombinedSection(),
		makeEnvVarsSection(),
		makeExamplesSection(),
	}
	fmt.Println(strings.Join(sections, "\n"))
}

func makeCombinedSection() string {
	usageWords := []string{"ez-utils", "population", "<directory>", "[options]"}
	styledUsageWords := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#C51010")).Render(usageWords[0]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render(usageWords[1]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCC00")).Render(usageWords[2]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFCC")).Render(usageWords[3]),
	}
	usage := headerStyle.Render("USAGE") + "\n\n" + strings.Join(styledUsageWords, " ")

	description := headerStyle.Render("DESCRIPTION") + "\n\n" +
		descStyle.Render("High-performance XML population processor that handles chunking, coordinate transformation (MTM8 to WGS84), spatial binning, and agent scaling.\nProcesses XML files from the input directory and outputs MATSim-compatible files in the same location.")

	flags := headerStyle.Render("FLAGS") + "\n\n" +
		descStyle.Render(
			flagStyle.Render("-d, --db")+": Enable PostgreSQL storage "+optionalStyle.Render("[optional]")+"\n"+
				flagStyle.Render("-s, --scale")+": Target population scale factor "+optionalStyle.Render("[optional]")+"\n  When not specified, automatically determines and produces up to 10 different scales (1% to 10%) based on population analysis\n"+
				flagStyle.Render("-c, --clean")+": Clean up temporary files after processing "+optionalStyle.Render("[optional]"),
		)

	return combinedSectionStyle.Render(
		usage + "\n\n" + description + "\n\n" + flags,
	)
}

func makeEnvVarsSection() string {
	envVars := []string{
		envVarStyle.Render("CHUNK_SIZE") + ": Defines the size of XML chunks for processing",
		envVarStyle.Render("DB_HOST") + ": Database host address",
		envVarStyle.Render("DB_PORT") + ": Database port",
		envVarStyle.Render("DB_NAME") + ": Database name",
		envVarStyle.Render("DB_USER") + ": Database username",
		envVarStyle.Render("DB_PASSWORD") + ": Database password",
	}

	return sectionStyle.Render(
		headerStyle.Render("ENVIRONMENT VARIABLES") + "\n\n" +
			descStyle.Render(strings.Join(envVars, "\n")),
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
