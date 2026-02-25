package form

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

// Component implements the form component
type Component struct {
	// Form component doesn't need internal state for this demo
}

// New creates a new form component instance
func New() *Component {
	return &Component{}
}

// submitHandler handles form submissions
func (c *Component) submitHandler(w http.ResponseWriter, r *http.Request) {
	// Check if this is a Datastar request
	if r.URL.Query().Get("datastar") != "" {
		// Read signals from request
		var signals map[string]interface{}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// In a real app, we would parse form data from signals
		// For demo, we'll simulate validation

		formData := DefaultForm()

		// Check for form field values in signals
		if name, ok := signals["name"].(string); ok && name != "" {
			formData = formData.WithFieldValue("name", name)
		} else {
			formData = formData.WithError("name", "Name is required")
		}

		if email, ok := signals["email"].(string); ok && email != "" {
			formData = formData.WithFieldValue("email", email)
			// Simple email validation
			if !strings.Contains(email, "@") {
				formData = formData.WithError("email", "Invalid email address")
			}
		} else {
			formData = formData.WithError("email", "Email is required")
		}

		if message, ok := signals["message"].(string); ok && message != "" {
			formData = formData.WithFieldValue("message", message)
		} else {
			formData = formData.WithError("message", "Message is required")
		}

		// Check if form is valid
		hasErrors := len(formData.Errors) > 0

		if !hasErrors {
			// Form is valid, mark as submitted
			formData = formData.WithSubmitted(true)
		}

		// Create SSE response
		sse := datastar.NewSSE(w, r)

		// Render component to string
		var buf strings.Builder
		formComponent := Form(formData)
		if err := formComponent.Render(r.Context(), &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Patch the form into the DOM
		sse.PatchElements(buf.String())

		// Update signals
		signalsUpdate := map[string]interface{}{
			"submitted": !hasErrors,
			"hasErrors": hasErrors,
		}
		jsonData, _ := json.Marshal(signalsUpdate)
		sse.PatchSignals(jsonData)

		return
	}

	// Fallback for non-Datastar requests (simple GET form)
	formData := DefaultForm()

	// Check for query parameters
	if name := r.URL.Query().Get("name"); name != "" {
		formData = formData.WithFieldValue("name", name)
	}
	if email := r.URL.Query().Get("email"); email != "" {
		formData = formData.WithFieldValue("email", email)
	}
	if message := r.URL.Query().Get("message"); message != "" {
		formData = formData.WithFieldValue("message", message)
	}

	// Check if this is a submission
	if r.URL.Query().Get("submit") == "true" {
		// Simple validation
		hasName := r.URL.Query().Get("name") != ""
		hasEmail := r.URL.Query().Get("email") != "" && strings.Contains(r.URL.Query().Get("email"), "@")
		hasMessage := r.URL.Query().Get("message") != ""

		if hasName && hasEmail && hasMessage {
			formData = formData.WithSubmitted(true)
		} else {
			if !hasName {
				formData = formData.WithError("name", "Name is required")
			}
			if !hasEmail {
				formData = formData.WithError("email", "Valid email is required")
			}
			if !hasMessage {
				formData = formData.WithError("message", "Message is required")
			}
		}
	}

	w.Header().Set("Content-Type", "text/html")
	formComponent := Form(formData)
	formComponent.Render(r.Context(), w)
}

// resetHandler resets the form
func (c *Component) resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("datastar") != "" {
		// Create SSE response
		sse := datastar.NewSSE(w, r)

		// Reset form to defaults
		formData := DefaultForm().ClearErrors()

		// Render component to string
		var buf strings.Builder
		formComponent := Form(formData)
		if err := formComponent.Render(r.Context(), &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Patch the form into the DOM
		sse.PatchElements(buf.String())

		// Update signals
		signalsUpdate := map[string]interface{}{
			"reset":     true,
			"submitted": false,
		}
		jsonData, _ := json.Marshal(signalsUpdate)
		sse.PatchSignals(jsonData)

		return
	}

	// Simple reset for non-Datastar
	formData := DefaultForm().ClearErrors()

	w.Header().Set("Content-Type", "text/html")
	formComponent := Form(formData)
	formComponent.Render(r.Context(), w)
}

// RegisterRoutes registers HTTP routes for the form component
func (c *Component) RegisterRoutes(r chi.Router) {
	r.Get("/submit", c.submitHandler)
	r.Get("/reset", c.resetHandler)
}

// RegisterStatic registers static asset routes for the form component
func (c *Component) RegisterStatic(r chi.Router) {
	// Form component CSS
	// r.Handle("/assets/*", http.FileServer(http.Dir("./components/form/assets")))
}
