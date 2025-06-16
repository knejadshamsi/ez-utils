package help

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	defaultHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#38B0DE")).
				Background(lipgloss.Color("#0A2A3F")).
				Padding(0, 1).
				Underline(true)

	defaultDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				PaddingLeft(2)

	defaultSectionStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				Padding(1).
				MarginBottom(1).
				Width(80)
)

func PrintDefaultHelp() {
	section := makeDefaultSection()
	fmt.Println(section)
}

func makeDefaultSection() string {
	// Style the usage section
	usageWords := []string{"ez-utils", "<command>", "[arguments]"}
	styledUsageWords := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#C51010")).Render(usageWords[0]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render(usageWords[1]),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFCC")).Render(usageWords[2]),
	}
	usage := defaultHeaderStyle.Render("USAGE") + "\n\n" + strings.Join(styledUsageWords, " ")

	// Style the available commands section
	commands := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render("scale") + defaultDescStyle.Render("         Scale down MATSim data"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render("network") + defaultDescStyle.Render("       Analyze network connections"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Render("create") + defaultDescStyle.Render("        Create new simulation components"),
	}
	availableCommands := defaultHeaderStyle.Render("AVAILABLE COMMANDS") + "\n\n" +
		strings.Join(commands, "\n")

	// Style the help usage section
	helpUsage := defaultHeaderStyle.Render("HELP") + "\n\n" +
		defaultDescStyle.Render("For detailed help on each command, use:") + "\n" +
		defaultDescStyle.Render("  <command> --help")

	// Style the example section
	example := defaultHeaderStyle.Render("EXAMPLES") + "\n\n" +
		defaultDescStyle.Render("  ez-utils scale population --help") + "\n" +
		defaultDescStyle.Render("  ez-utils scale population population.xml")

	// Combine all sections
	return defaultSectionStyle.Render(
		strings.Join([]string{
			usage,
			availableCommands,
			helpUsage,
			example,
		}, "\n\n"),
	)
}
