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
		makeNetworkEnvVarsSection(),
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

func makeNetworkEnvVarsSection() string {
	envVars := []string{
		networkFlagStyle.Render("DB_HOST") + ": Database host address",
		networkFlagStyle.Render("DB_PORT") + ": Database port",
		networkFlagStyle.Render("DB_NAME") + ": Database name",
		networkFlagStyle.Render("DB_USER") + ": Database username",
		networkFlagStyle.Render("DB_PASSWORD") + ": Database password",
	}

	return networkSectionStyle.Render(
		networkHeaderStyle.Render("ENVIRONMENT VARIABLES") + "\n\n" +
			networkDescStyle.Render(strings.Join(envVars, "\n")),
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
