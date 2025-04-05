package help

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	networkHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.Color("#39A7FF"))

	networkFlagStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFE0"))

	networkDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Padding(1)

	networkSectionStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				PaddingBottom(1)
)

func PrintNetworkHelp() {
	usageSection := networkSectionStyle.Render(
		networkHeaderStyle.Render("USAGE") + "\n" +
			networkDescStyle.Render("ez-utils network <directory>"))

	descriptionSection := networkSectionStyle.Render(
		networkHeaderStyle.Render("DESCRIPTION") + "\n" +
			networkDescStyle.Render("High-performance network analysis pipeline that processes agent connections across multiple scales (1-10). "+
				"Handles XML chunk validation, agent/link indexing, multi-scale processing, and connection analysis. "+
				"Supports streaming operations for memory efficiency and automatically consolidates results by scale."))

	envVarsContent := "DB_HOST: Database host address\n" +
		"DB_PORT: Database port\n" +
		"DB_NAME: Database name\n" +
		"DB_USER: Database username\n" +
		"DB_PASSWORD: Database password"

	envVarsSection := networkSectionStyle.Render(
		networkHeaderStyle.Render("ENVIRONMENT VARIABLES") + "\n" +
			networkDescStyle.Render(envVarsContent))

	flagsContent := networkFlagStyle.Render("-d, --db") + ": Enable PostgreSQL storage (optional)\n" +
		networkFlagStyle.Render("-c, --clean") + ": Clean up temporary files after processing (optional)"

	flagsSection := networkSectionStyle.Render(
		networkHeaderStyle.Render("FLAGS") + "\n" +
			networkDescStyle.Render(flagsContent))

	examplesContent := "ez-utils network ./my-network\n" +
		"ez-utils network ./my-network -d\n" +
		"ez-utils network ./my-network --db --clean"

	examplesSection := networkSectionStyle.Render(
		networkHeaderStyle.Render("EXAMPLES") + "\n" +
			networkDescStyle.Render(examplesContent))

	fmt.Print(
		usageSection + "\n\n" +
			descriptionSection + "\n\n" +
			envVarsSection + "\n\n" +
			flagsSection + "\n\n" +
			examplesSection + "\n")
}
