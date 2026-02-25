package form

import (
	"context"
	"strings"

	"github.com/a-h/templ"
)

func renderComponentToString(c templ.Component) (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := c.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// FieldType represents the type of form field
type FieldType string

const (
	FieldText     FieldType = "text"
	FieldEmail    FieldType = "email"
	FieldPassword FieldType = "password"
	FieldNumber   FieldType = "number"
	FieldSelect   FieldType = "select"
	FieldTextarea FieldType = "textarea"
	FieldCheckbox FieldType = "checkbox"
)

// FormField represents a single form field
type FormField struct {
	ID          string
	Name        string
	Label       string
	Type        FieldType
	Value       string
	Placeholder string
	Required    bool
	Disabled    bool
	Options     []SelectOption // For select fields
	Error       string
}

// SelectOption represents an option for select fields
type SelectOption struct {
	Value string
	Label string
}

// FormData represents a form component's data
type FormData struct {
	ID        string
	Title     string
	Action    string
	Method    string // "GET", "POST"
	Fields    []FormField
	Submitted bool
	Errors    map[string]string
}

// DefaultForm creates a form with default values
func DefaultForm() FormData {
	return FormData{
		ID:     "demo-form",
		Title:  "Contact Form",
		Action: "/api/form/submit",
		Method: "POST",
		Fields: []FormField{
			{
				ID:          "name",
				Name:        "name",
				Label:       "Name",
				Type:        FieldText,
				Value:       "",
				Placeholder: "Enter your name",
				Required:    true,
			},
			{
				ID:          "email",
				Name:        "email",
				Label:       "Email",
				Type:        FieldEmail,
				Value:       "",
				Placeholder: "Enter your email",
				Required:    true,
			},
			{
				ID:          "message",
				Name:        "message",
				Label:       "Message",
				Type:        FieldTextarea,
				Value:       "",
				Placeholder: "Enter your message",
				Required:    true,
			},
			{
				ID:       "newsletter",
				Name:     "newsletter",
				Label:    "Subscribe to newsletter",
				Type:     FieldCheckbox,
				Value:    "true",
				Required: false,
			},
		},
		Submitted: false,
		Errors:    make(map[string]string),
	}
}

// GenerateHTML generates the HTML for a form
func (f FormData) GenerateHTML() string {
	component := FormComponent(f)
	html, err := renderComponentToString(component)
	if err != nil {
		return ""
	}
	return html
}

// WithFieldValue updates a field's value
func (f FormData) WithFieldValue(fieldID, value string) FormData {
	for i, field := range f.Fields {
		if field.ID == fieldID {
			f.Fields[i].Value = value
			break
		}
	}
	return f
}

// WithSubmitted sets the submitted state
func (f FormData) WithSubmitted(submitted bool) FormData {
	f.Submitted = submitted
	return f
}

// WithError adds an error to a field
func (f FormData) WithError(fieldID, error string) FormData {
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Errors[fieldID] = error

	// Also update the field's error
	for i, field := range f.Fields {
		if field.ID == fieldID {
			f.Fields[i].Error = error
			break
		}
	}

	return f
}

// ClearErrors clears all errors
func (f FormData) ClearErrors() FormData {
	f.Errors = make(map[string]string)
	for i := range f.Fields {
		f.Fields[i].Error = ""
	}
	return f
}
