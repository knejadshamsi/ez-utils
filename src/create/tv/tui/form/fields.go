package form

import (
	"fmt"
	"strconv"
	"github.com/charmbracelet/bubbles/textinput"
)

// AddNumberField adds a number field
func (f *Form) AddNumberField(label string, placeholder string, required bool) *Form {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(placeholder) // Set actual value, not just placeholder
	ti.CharLimit = 20
	
	field := FormField{
		label:     label,
		input:     ti,
		validate:  validatePositiveIntForm,
		required:  required,
		fieldType: "number",
	}
	
	f.fields = append(f.fields, field)
	return f
}

// AddCapacityField adds a capacity field that allows 0 values
func (f *Form) AddCapacityField(label string, placeholder string, required bool) *Form {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(placeholder) // Set actual value, not just placeholder
	ti.CharLimit = 20
	
	field := FormField{
		label:     label,
		input:     ti,
		validate:  validateNonNegativeIntForm,
		required:  required,
		fieldType: "number",
	}
	
	f.fields = append(f.fields, field)
	return f
}

// AddCapacityFieldWithValue adds a capacity field with an initial value
func (f *Form) AddCapacityFieldWithValue(label string, initialValue string, placeholder string, required bool) *Form {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(initialValue)
	ti.CharLimit = 20
	
	field := FormField{
		label:     label,
		input:     ti,
		validate:  validateNonNegativeIntForm,
		required:  required,
		fieldType: "number",
	}
	
	f.fields = append(f.fields, field)
	return f
}

// AddFloatField adds a float number field
func (f *Form) AddFloatField(label string, placeholder string, required bool) *Form {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(placeholder) // Set actual value, not just placeholder
	ti.CharLimit = 20
	
	field := FormField{
		label:     label,
		input:     ti,
		validate:  validatePositiveFloatForm,
		required:  required,
		fieldType: "float",
	}
	
	f.fields = append(f.fields, field)
	return f
}

// AddSelectField adds a select field
func (f *Form) AddSelectField(label string, options []string, required bool) *Form {
	ti := textinput.New()
	ti.Placeholder = "Press Enter to select"
	ti.CharLimit = 0 // Read-only
	
	field := FormField{
		label:       label,
		input:       ti,
		required:    required,
		fieldType:   "select",
		options:     options,
		optionIndex: 0,
	}
	
	// Set initial value if options exist
	if len(options) > 0 {
		field.input.SetValue(options[0])
	}
	
	f.fields = append(f.fields, field)
	return f
}

// Validation helpers

func validatePositiveIntForm(s string) error {
	if s == "" {
		return nil
	}
	
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("must be a valid integer")
	}
	
	if n <= 0 {
		return fmt.Errorf("must be greater than 0")
	}
	
	return nil
}

func validatePositiveFloatForm(s string) error {
	if s == "" {
		return nil
	}
	
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("must be a valid number")
	}
	
	if n <= 0 {
		return fmt.Errorf("must be greater than 0")
	}
	
	return nil
}

func validateNonNegativeIntForm(s string) error {
	if s == "" {
		return nil
	}
	
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("must be a valid integer")
	}
	
	if n < 0 {
		return fmt.Errorf("must be 0 or greater")
	}
	
	return nil
}