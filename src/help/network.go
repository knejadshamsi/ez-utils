package help

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	networkHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#38B0DE")).
				Background(lipgloss.Color("#0A2A3F")).
				Padding(0, 1).
				Underline(true)

	networkFlagStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Background(lipgloss.Color("#0A2A3F")).
				Padding(0, 1)

	networkDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				PaddingLeft(2)

	networkSectionStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				Padding(1).
				MarginBottom(1).
				Width(80)

	networkCombinedSectionStyle = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder()).
					Padding(1).
					MarginBottom(1).
					Width(80)

	networkOptionalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4040"))
)

func PrintNetworkHelp() {
	sections := []string{
		makeNetworkCombinedSection(),
		makeNetworkConfigSection(),
		makeNetworkExamplesSection(),
	}
	fmt.Println(strings.Join(sections, "\n"))
}

func makeNetworkCombinedSection() string {
	usageWords := []string{"ez-utils", "network", "<directory>"}
	styledUsageWords := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#C51010")).Render(usageWords[0]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render(usageWords[1]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCC00")).Render(usageWords[2]),
	}
	usage := networkHeaderStyle.Render("USAGE") + "\n\n" + strings.Join(styledUsageWords, " ")

	description := networkHeaderStyle.Render("DESCRIPTION") + "\n\n" +
		networkDescStyle.Render("High-performance network analysis pipeline that processes agent connections across multiple scales (1-10). "+
			"Handles XML chunk validation, agent/link indexing, multi-scale processing, and connection analysis. "+
			"Supports streaming operations for memory efficiency and automatically consolidates results by scale.")

	flags := networkHeaderStyle.Render("FLAGS") + "\n\n" +
		networkDescStyle.Render(
			networkFlagStyle.Render("-d, --db")+": Enable PostgreSQL storage "+networkOptionalStyle.Render("[optional]")+"\n"+
				networkFlagStyle.Render("-c, --clean")+": Clean up temporary files after processing "+networkOptionalStyle.Render("[optional]"),
		)

	return networkCombinedSectionStyle.Render(
		usage + "\n\n" + description + "\n\n" + flags,
	)
}

func makeNetworkConfigSection() string {
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
		"Use " + networkFlagStyle.Render("--new-config") + " to create a new config file interactively",
		"",
		"Database settings (required for network analysis):",
		networkFlagStyle.Render("database.host") + ": Database host address",
		networkFlagStyle.Render("database.port") + ": Database port (default: 5432)",
		networkFlagStyle.Render("database.name") + ": Database name",
		networkFlagStyle.Render("database.user") + ": Database username",
		networkFlagStyle.Render("database.password") + ": Database password",
	}

	return networkSectionStyle.Render(
		networkHeaderStyle.Render("CONFIGURATION") + "\n\n" +
			networkDescStyle.Render(strings.Join(configInfo, "\n")),
	)
}

func makeNetworkExamplesSection() string {
	examples := []string{
		"ez-utils network ./my-network",
		"ez-utils network ./my-network -d",
		"ez-utils network ./my-network --db --clean",
	}

	return networkSectionStyle.Render(
		networkHeaderStyle.Render("EXAMPLES") + "\n\n" +
			networkDescStyle.Render(strings.Join(examples, "\n")),
	)
}
