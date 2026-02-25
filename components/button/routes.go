package button

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

// Component implements the button component
type Component struct {
	// Button component doesn't need internal state for this demo
}

// New creates a new button component instance
func New() *Component {
	return &Component{}
}

// clickHandler handles button click events
func (c *Component) clickHandler(w http.ResponseWriter, r *http.Request) {
	// Check if this is a Datastar request
	if r.URL.Query().Get("datastar") != "" {
		// Read signals from request
		var signals map[string]interface{}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get current click count from signals or default to 0
		clickCount := 0
		if count, ok := signals["clickCount"].(float64); ok {
			clickCount = int(count)
		}

		// Increment click count
		clickCount++

		// Create updated button data
		data := DefaultButton().
			WithLabel(fmt.Sprintf("Clicked %d times", clickCount)).
			WithClickCount(clickCount)

		// If click count is high, change variant for fun
		if clickCount >= 5 {
			data = data.WithVariant("danger")
		} else if clickCount >= 3 {
			data = data.WithVariant("success")
		}

		// Create SSE response
		sse := datastar.NewSSE(w, r)

		// Render component to string
		var buf strings.Builder
		buttonComponent := Button(data)
		if err := buttonComponent.Render(r.Context(), &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Patch the button into the DOM
		sse.PatchElements(buf.String())

		// Also update signals with new click count
		signalsUpdate := map[string]interface{}{
			"clickCount": clickCount,
			"lastClick":  time.Now().Format(time.RFC3339),
		}
		jsonData, _ := json.Marshal(signalsUpdate)
		sse.PatchSignals(jsonData)

		return
	}

	// Fallback for non-Datastar requests
	clickCountStr := r.URL.Query().Get("count")
	clickCount := 0
	if clickCountStr != "" {
		if count, err := strconv.Atoi(clickCountStr); err == nil {
			clickCount = count
		}
	}

	clickCount++
	data := DefaultButton().
		WithLabel(fmt.Sprintf("Clicked %d times", clickCount)).
		WithClickCount(clickCount)

	w.Header().Set("Content-Type", "text/html")
	buttonComponent := Button(data)
	buttonComponent.Render(r.Context(), w)
}

// toggleHandler toggles button state (enabled/disabled, loading/not loading)
func (c *Component) toggleHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("datastar") != "" {
		var signals map[string]interface{}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get current state from signals
		disabled := false
		loading := false
		clickCount := 0

		if dis, ok := signals["disabled"].(bool); ok {
			disabled = dis
		}
		if load, ok := signals["loading"].(bool); ok {
			loading = load
		}
		if count, ok := signals["clickCount"].(float64); ok {
			clickCount = int(count)
		}

		// Toggle disabled state
		disabled = !disabled

		// If enabling, set loading for 2 seconds
		if !disabled {
			loading = true
		}

		data := DefaultButton().
			WithLabel("Processing...").
			WithDisabled(disabled).
			WithLoading(loading).
			WithClickCount(clickCount)

		sse := datastar.NewSSE(w, r)

		var buf strings.Builder
		buttonComponent := Button(data)
		if err := buttonComponent.Render(r.Context(), &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sse.PatchElements(buf.String())

		// Update signals
		signalsUpdate := map[string]interface{}{
			"disabled":   disabled,
			"loading":    loading,
			"clickCount": clickCount,
		}
		jsonData, _ := json.Marshal(signalsUpdate)
		sse.PatchSignals(jsonData)

		return
	}

	// Simple toggle for non-Datastar
	disabled := r.URL.Query().Get("disabled") != "true"
	data := DefaultButton().
		WithDisabled(disabled).
		WithLabel(func() string {
			if disabled {
				return "Disabled Button"
			}
			return "Enabled Button"
		}())

	w.Header().Set("Content-Type", "text/html")
	buttonComponent := Button(data)
	buttonComponent.Render(r.Context(), w)
}

// RegisterRoutes registers HTTP routes for the button component
func (c *Component) RegisterRoutes(r chi.Router) {
	r.Get("/click", c.clickHandler)
	r.Get("/toggle", c.toggleHandler)
}

// RegisterStatic registers static asset routes for the button component
func (c *Component) RegisterStatic(r chi.Router) {
	// Button component CSS
	// r.Handle("/assets/*", http.FileServer(http.Dir("./components/button/assets")))
}
