package help

import (
	"encoding/json"
	"embed"
	"fmt"
	"strings"

	"ez-utils/config"

	"github.com/charmbracelet/lipgloss"
)

//go:embed help_data.json
var helpDataFS embed.FS

// --- Data Structures for help_data.json ---

type HelpData struct {
	En HelpContent `json:"en"`
	Fr HelpContent `json:"fr"`
}

type HelpContent struct {
	Default    CommandHelp `json:"default"`
	Scale      CommandHelp `json:"scale"`
	Create     CommandHelp `json:"create"`
	Edit       CommandHelp `json:"edit"`
	Population CommandHelp `json:"population"`
	Pt         CommandHelp `json:"pt"`
}

type CommandHelp struct {
	Usage             string        `json:"usage"`
	DescriptionHeader string        `json:"description_header"`
	Description       string        `json:"description"`
	CommandsHeader    string        `json:"commands_header"`
	Commands          []CommandDesc `json:"commands"`
	HelpHeader        string        `json:"help_header"`
	HelpDesc          string        `json:"help_desc"`
	HelpUsage         string        `json:"help_usage"`
	FlagsHeader       string        `json:"flags_header"`
	Flags             []FlagDesc    `json:"flags"`
	ConfigHeader      string        `json:"config_header"`
	ConfigInfo        []string      `json:"config_info"`
	TimePeriodsHeader string        `json:"time_periods_header"`
	TimePeriods       []CommandDesc `json:"time_periods"`
	ExamplesHeader    string        `json:"examples_header"`
	Examples          []string      `json:"examples"`
}

type CommandDesc struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type FlagDesc struct {
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Optional bool   `json:"optional"`
}

// --- Centralized Lipgloss Styles ---

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#38B0DE")).
			Background(lipgloss.Color("#0A2A3F")).
			Padding(0, 1).
			Underline(true)

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF9900"))

	argStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCC00"))

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
)

var helpData *HelpData

// LoadHelpData reads and parses the help_data.json file.
func LoadHelpData() error {
	data, err := helpDataFS.ReadFile("help_data.json")
	if err != nil {
		return fmt.Errorf("failed to read embedded help_data.json: %w", err)
	}

	var loadedData HelpData
	if err := json.Unmarshal(data, &loadedData); err != nil {
		return fmt.Errorf("failed to parse embedded help_data.json: %w", err)
	}
	helpData = &loadedData
	return nil
}

// PrintHelp displays the help message for a specific command and subcommand.
func PrintHelp(command, subcommand string, appConfig *config.Config) {
	if helpData == nil {
		if err := LoadHelpData(); err != nil {
			fmt.Println("Error loading help data:", err)
			return
		}
	}

	var content HelpContent
	switch appConfig.Language {
	case "fr":
		content = helpData.Fr
	default:
		content = helpData.En
	}

	var commandHelp CommandHelp
	switch command {
	case "scale":
		if subcommand == "population" {
			commandHelp = content.Population
		} else {
			commandHelp = content.Scale
		}
	case "create":
		if subcommand == "pt" {
			commandHelp = content.Pt
		} else {
			commandHelp = content.Create
		}
	case "edit":
		commandHelp = content.Edit
	default:
		commandHelp = content.Default
	}

	fmt.Println(buildHelpMessage(commandHelp))
}

// buildHelpMessage constructs the final formatted help string.
func buildHelpMessage(ch CommandHelp) string {
	var sections []string

	if ch.Usage != "" {
		sections = append(sections, headerStyle.Render("USAGE")+"\n\n"+descStyle.Render(ch.Usage))
	}

	if ch.Description != "" {
		sections = append(sections, headerStyle.Render(strings.ToUpper(ch.DescriptionHeader))+"\n\n"+descStyle.Render(ch.Description))
	}

	if len(ch.Commands) > 0 {
		var cmdLines []string
		for _, cmd := range ch.Commands {
			cmdLines = append(cmdLines, commandStyle.Render(cmd.Name)+descStyle.Render(" "+cmd.Desc))
		}
		sections = append(sections, headerStyle.Render(strings.ToUpper(ch.CommandsHeader))+"\n\n"+strings.Join(cmdLines, "\n"))
	}

	if len(ch.Flags) > 0 {
		var flagLines []string
		for _, flag := range ch.Flags {
			line := commandStyle.Render(flag.Name) + descStyle.Render(" "+flag.Desc)
			if flag.Optional {
				line += " " + optionalStyle.Render("[optional]")
			}
			flagLines = append(flagLines, line)
		}
		sections = append(sections, headerStyle.Render(strings.ToUpper(ch.FlagsHeader))+"\n\n"+strings.Join(flagLines, "\n"))
	}

	if len(ch.ConfigInfo) > 0 {
		sections = append(sections, headerStyle.Render(strings.ToUpper(ch.ConfigHeader))+"\n\n"+descStyle.Render(strings.Join(ch.ConfigInfo, "\n")))
	}

	if len(ch.TimePeriods) > 0 {
		var periodLines []string
		for _, period := range ch.TimePeriods {
			periodLines = append(periodLines, commandStyle.Render(period.Name)+descStyle.Render(" "+period.Desc))
		}
		sections = append(sections, headerStyle.Render(strings.ToUpper(ch.TimePeriodsHeader))+"\n\n"+strings.Join(periodLines, "\n"))
	}

	if ch.HelpDesc != "" {
		helpStr := headerStyle.Render(strings.ToUpper(ch.HelpHeader)) + "\n\n" + descStyle.Render(ch.HelpDesc) + "\n" + descStyle.Render("  "+ch.HelpUsage)
		sections = append(sections, helpStr)
	}

	if len(ch.Examples) > 0 {
		sections = append(sections, headerStyle.Render(strings.ToUpper(ch.ExamplesHeader))+"\n\n"+descStyle.Render(strings.Join(ch.Examples, "\n")))
	}

	return sectionStyle.Render(strings.Join(sections, "\n\n"))
}
