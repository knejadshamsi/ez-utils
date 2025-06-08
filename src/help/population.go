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
		makeConfigSection(),
		makeExamplesSection(),
	}
	fmt.Println(strings.Join(sections, "\n"))
}

func makeCombinedSection() string {
	usageWords := []string{"ez-utils", "scale", "population", "<population-file>", "[options]"}
	styledUsageWords := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#C51010")).Render(usageWords[0]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render(usageWords[1]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render(usageWords[2]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCC00")).Render(usageWords[3]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFCC")).Render(usageWords[4]),
	}
	usage := headerStyle.Render("USAGE") + "\n\n" + strings.Join(styledUsageWords, " ")

	description := headerStyle.Render("DESCRIPTION") + "\n\n" +
		descStyle.Render("High-performance XML population processor that handles chunking, coordinate transformation (MTM8 to WGS84), spatial binning, and agent scaling.\nProcesses the specified population XML file and outputs MATSim-compatible scaled files.")

	flags := headerStyle.Render("FLAGS") + "\n\n" +
		descStyle.Render(
			flagStyle.Render("--db")+": Enable PostgreSQL storage "+optionalStyle.Render("[optional]")+"\n"+
				flagStyle.Render("-s, --scale")+": Target population scale factor "+optionalStyle.Render("[optional]")+"\n  When not specified, automatically determines and produces up to 10 different scales (1% to 10%) based on population analysis\n"+
				flagStyle.Render("-c, --clean")+": Clean up temporary files after processing "+optionalStyle.Render("[optional]"),
		)

	return combinedSectionStyle.Render(
		usage + "\n\n" + description + "\n\n" + flags,
	)
}

func makeConfigSection() string {
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
		"Use " + flagStyle.Render("--new-config") + " to create a new config file interactively",
		"",
		"Key settings for population processing:",
		envVarStyle.Render("chunk_size") + ": Size of XML chunks for processing (default: 2000)",
		envVarStyle.Render("output_dir") + ": Directory for output files (default: output)",
		envVarStyle.Render("language") + ": UI language - en or fr (default: en)",
		"",
		"Database settings (required when using --db):",
		envVarStyle.Render("database.host") + ": Database host address",
		envVarStyle.Render("database.port") + ": Database port (default: 5432)",
		envVarStyle.Render("database.name") + ": Database name",
		envVarStyle.Render("database.user") + ": Database username",
		envVarStyle.Render("database.password") + ": Database password",
	}

	return sectionStyle.Render(
		headerStyle.Render("CONFIGURATION") + "\n\n" +
			descStyle.Render(strings.Join(configInfo, "\n")),
	)
}

func makeExamplesSection() string {
	examples := []string{
		"ez-utils scale population population.xml",
		"ez-utils scale population population.xml --db",
		"ez-utils scale population population.xml --db -s 0.1 --clean",
	}

	return sectionStyle.Render(
		headerStyle.Render("EXAMPLES") + "\n\n" +
			descStyle.Render(strings.Join(examples, "\n")),
	)
}
