package form

import (
	"fmt"
	"log"
	"os"
	"strings"
	
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormSubmittedMsg is sent when a form is successfully submitted
type FormSubmittedMsg struct {
	Values map[string]string
}

// ModalCloseMsg is sent when a modal should be closed
type ModalCloseMsg struct{}

// Init initializes the form
func (f *Form) Init() tea.Cmd {
	// Force all text inputs to properly reset their cursor and state
	for i := range f.fields {
		if f.fields[i].fieldType != "select" {
			// Reset cursor to end of text
			f.fields[i].input.SetCursor(len(f.fields[i].input.Value()))
		}
	}
	
	// Focus first field
	if len(f.fields) > 0 {
		f.fields[0].input.Focus()
		return textinput.Blink
	}
	return nil
}

// Update handles form updates
func (f *Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			f.validateCurrentField()
			f.nextField()
		case "shift+tab", "up":
			f.validateCurrentField()
			f.prevField()
		case "enter":
			// LOG FORM-1: Enter key pressed
			logFile, err := os.OpenFile("vt-edit-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				logger := log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
				logger.Printf("[FORM-1] Enter key pressed in form")
				logger.Printf("[FORM-2] Form title: %s", f.title)
				logFile.Close()
			}
			
			// Always submit on Enter
			if f.validateAll() {
				// LOG FORM-3: Validation passed
				logFile, _ := os.OpenFile("vt-edit-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				logger := log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
				logger.Printf("[FORM-3] Validation PASSED")
				logger.Printf("[FORM-4] onSubmit handler exists: %v", f.onSubmit != nil)
				logFile.Close()
				
				if f.onSubmit != nil {
					values := f.getValues()
					
					// LOG FORM-5: Values collected
					logFile, _ := os.OpenFile("vt-edit-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					logger := log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
					logger.Printf("[FORM-5] Form values collected: %+v", values)
					logger.Printf("[FORM-6] Calling onSubmit handler...")
					logFile.Close()
					
					if err := f.onSubmit(values); err != nil {
						// LOG FORM-7: Submit error
						logFile, _ := os.OpenFile("vt-edit-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						logger := log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
						logger.Printf("[FORM-7] onSubmit returned ERROR: %v", err)
						logFile.Close()
						
						// Show error
						f.errors[-1] = err.Error()
					} else {
						// LOG FORM-8: Submit success
						logFile, _ := os.OpenFile("vt-edit-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						logger := log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
						logger.Printf("[FORM-8] onSubmit returned SUCCESS")
						logger.Printf("[FORM-9] Sending FormSubmittedMsg")
						logFile.Close()
						
						// Form submitted successfully - send message to close modal
						delete(f.errors, -1)
						return *f, func() tea.Msg {
							return FormSubmittedMsg{Values: values}
						}
					}
				}
			} else {
				// LOG FORM-10: Validation failed
				logFile, _ := os.OpenFile("vt-edit-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				logger := log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
				logger.Printf("[FORM-10] Validation FAILED")
				for idx, err := range f.errors {
					logger.Printf("[FORM-11] Error at field %d: %s", idx, err)
				}
				logFile.Close()
			}
		case " ", "space":
			// Use space to cycle select options
			if f.fields[f.focusIndex].fieldType == "select" {
				f.cycleSelectOption(f.focusIndex)
			}
		case "left":
			if f.fields[f.focusIndex].fieldType == "select" {
				f.cycleSelectOption(f.focusIndex)
			}
		case "right":
			if f.fields[f.focusIndex].fieldType == "select" {
				f.cycleSelectOption(f.focusIndex)
			}
		}
	}

	// Update current field
	if f.focusIndex < len(f.fields) {
		field := &f.fields[f.focusIndex]
		if field.fieldType != "select" {
			newInput, cmd := field.input.Update(msg)
			field.input = newInput
			cmds = append(cmds, cmd)
			
			// Validate on change
			if field.validate != nil {
				value := field.input.Value()
				if err := field.validate(value); err != nil {
					f.errors[f.focusIndex] = err.Error()
				} else {
					delete(f.errors, f.focusIndex)
				}
			}
		}
	}

	return *f, tea.Batch(cmds...)
}

// View renders the form
func (f *Form) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFF")).
		MarginBottom(1)
	
	b.WriteString(titleStyle.Render(f.title))
	b.WriteString("\n")
	
	// Description
	if f.description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAA")).
			MarginBottom(1)
		b.WriteString(descStyle.Render(f.description))
		b.WriteString("\n\n")
	}

	// Fields
	for i, field := range f.fields {
		// Label
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF"))
		if i == f.focusIndex {
			labelStyle = labelStyle.Bold(true).Foreground(lipgloss.Color("#7571F9"))
		}
		
		label := field.label
		if field.required {
			label += " *"
		}
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")
		
		// Input
		inputView := field.input.View()
		if field.fieldType == "select" && i == f.focusIndex {
			// Show select options
			inputView = fmt.Sprintf("← %s →", field.input.Value())
		}
		
		b.WriteString(inputView)
		b.WriteString("\n")
		
		// Error
		if err, exists := f.errors[i]; exists {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F00")).
				Italic(true)
			b.WriteString(errorStyle.Render("  " + err))
			b.WriteString("\n")
		}
		
		b.WriteString("\n")
	}
	
	// Submit hint
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Italic(true)
	b.WriteString(hintStyle.Render("Press Enter to submit, Tab/Shift+Tab to navigate"))
	
	// General error
	if err, exists := f.errors[-1]; exists {
		b.WriteString("\n\n")
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F00")).
			Bold(true)
		b.WriteString(errorStyle.Render("Error: " + err))
	}

	return b.String()
}

// nextField moves focus to the next field
func (f *Form) nextField() {
	if len(f.fields) > 0 {
		f.fields[f.focusIndex].input.Blur()
		f.focusIndex = (f.focusIndex + 1) % len(f.fields)
		f.fields[f.focusIndex].input.Focus()
	}
}

// prevField moves focus to the previous field
func (f *Form) prevField() {
	if len(f.fields) > 0 {
		f.fields[f.focusIndex].input.Blur()
		f.focusIndex = (f.focusIndex - 1 + len(f.fields)) % len(f.fields)
		f.fields[f.focusIndex].input.Focus()
	}
}

// cycleSelectOption cycles through select field options
func (f *Form) cycleSelectOption(index int) {
	if index < len(f.fields) && f.fields[index].fieldType == "select" {
		field := &f.fields[index]
		if len(field.options) > 0 {
			field.optionIndex = (field.optionIndex + 1) % len(field.options)
			field.input.SetValue(field.options[field.optionIndex])
		}
	}
}

// validateAll validates all fields and returns true if valid
func (f *Form) validateAll() bool {
	valid := true
	for i, field := range f.fields {
		value := field.input.Value()
		
		// Check required fields
		if field.required && value == "" {
			f.errors[i] = "This field is required"
			valid = false
			continue
		}
		
		// Run custom validation
		if field.validate != nil {
			if err := field.validate(value); err != nil {
				f.errors[i] = err.Error()
				valid = false
				continue
			}
		}
		
		// Clear any existing error for this field
		delete(f.errors, i)
	}
	return valid
}

// validateCurrentField validates the currently focused field
func (f *Form) validateCurrentField() {
	if f.focusIndex < len(f.fields) {
		field := f.fields[f.focusIndex]
		value := field.input.Value()
		
		// Check if required and empty
		if field.required && value == "" {
			f.errors[f.focusIndex] = "This field is required"
			return
		}
		
		// Run validation if present
		if field.validate != nil {
			if err := field.validate(value); err != nil {
				f.errors[f.focusIndex] = err.Error()
				return
			}
		}
		
		// Clear error if validation passes
		delete(f.errors, f.focusIndex)
	}
}

// getValues returns all field values as a map
func (f *Form) getValues() map[string]string {
	values := make(map[string]string)
	for _, field := range f.fields {
		values[field.label] = field.input.Value()
	}
	return values
}