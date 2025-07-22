package form

import (
	"github.com/charmbracelet/bubbles/textinput"
)

// Styles placeholder - will be properly defined when model package is implemented
type Styles struct {
	// Placeholder for styles - will be moved to model package
}

// FormField represents a single form field
type FormField struct {
	label       string
	input       textinput.Model
	validate    func(string) error
	required    bool
	fieldType   string // "text", "number", "select"
	options     []string // for select fields
	optionIndex int
}

// Form represents a complete form with multiple fields
type Form struct {
	title        string
	description  string
	fields       []FormField
	focusIndex   int
	errors       map[int]string
	width        int
	height       int
	onSubmit     func(map[string]string) error
	submitLabel  string
	styles       *Styles
}

// NewForm creates a new form
func NewForm(title, description string, styles *Styles) *Form {
	return &Form{
		title:       title,
		description: description,
		fields:      []FormField{},
		errors:      make(map[int]string),
		submitLabel: "Submit",
		styles:      styles,
	}
}

// AddField adds a field to the form
func (f *Form) AddField(label string, placeholder string, required bool, validate func(string) error) *Form {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 100
	
	field := FormField{
		label:     label,
		input:     ti,
		validate:  validate,
		required:  required,
		fieldType: "text",
	}
	
	f.fields = append(f.fields, field)
	return f
}

// AddFieldWithValue adds a field with an initial value
func (f *Form) AddFieldWithValue(label string, initialValue string, placeholder string, required bool, validate func(string) error) *Form {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(initialValue)
	ti.CharLimit = 100
	
	field := FormField{
		label:     label,
		input:     ti,
		validate:  validate,
		required:  required,
		fieldType: "text",
	}
	
	f.fields = append(f.fields, field)
	return f
}

// SetSubmitHandler sets the submit handler function
func (f *Form) SetSubmitHandler(handler func(map[string]string) error) *Form {
	f.onSubmit = handler
	return f
}