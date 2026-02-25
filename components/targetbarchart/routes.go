package targetbarchart

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

var barColors = []string{
	"#4f46e5", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6",
	"#06b6d4", "#84cc16", "#f97316", "#ec4899", "#14b8a6",
}

var productLabels = []string{
	"Widget A", "Widget B", "Widget C", "Widget D", "Widget E",
	"Widget F", "Widget G", "Widget H", "Widget I", "Widget J",
}

// Component implements the target bar chart component
type Component struct {
	rand *rand.Rand
	data TargetBarChartData // current chart data
}

// New creates a new target bar chart component instance
func New() *Component {
	comp := &Component{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	comp.data = comp.GenerateRandomData()
	return comp
}

// GenerateRandomData creates random target bar chart data
func (c *Component) GenerateRandomData() TargetBarChartData {
	// Generate 5-8 random bars
	numBars := c.rand.Intn(4) + 5

	bars := make([]BarData, numBars)

	for i := 0; i < numBars; i++ {
		target := c.rand.Intn(20) + 5        // 5-24 sections
		completed := c.rand.Intn(target + 1) // 0 to target
		bars[i] = BarData{
			Label:     productLabels[i%len(productLabels)],
			Target:    target,
			Completed: completed,
			Color:     barColors[i%len(barColors)],
		}
	}

	data := DefaultTargetBarChart()
	data.Bars = bars
	data.SVG = data.GenerateSVGString()
	data.HTML = data.GenerateHTML()
	return data
}

// updateBar increments or decrements completed count for a specific bar
func (c *Component) updateBar(barIndex int, delta int) bool {
	if barIndex < 0 || barIndex >= len(c.data.Bars) {
		return false
	}
	bar := &c.data.Bars[barIndex]
	newCompleted := bar.Completed + delta
	if newCompleted < 0 {
		newCompleted = 0
	}
	if newCompleted > bar.Target {
		newCompleted = bar.Target
	}
	bar.Completed = newCompleted
	// Regenerate SVG and HTML
	c.data.SVG = c.data.GenerateSVGString()
	c.data.HTML = c.data.GenerateHTML()
	return true
}

// randomizeHandler handles requests to randomize the target bar chart
func (c *Component) randomizeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if this is a Datastar request by looking for the datastar query param
	if r.URL.Query().Get("datastar") != "" {
		// Read signals from request
		var signals map[string]interface{}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate new random data
		c.data = c.GenerateRandomData()

		// Create the target bar chart component
		chartComponent := TargetBarChart(c.data)

		// Create SSE response
		sse := datastar.NewSSE(w, r)

		// Render component to string
		var buf strings.Builder
		if err := chartComponent.Render(r.Context(), &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Patch the chart into the DOM (morphing into element with id="target-bar-chart")
		sse.PatchElements(buf.String())

		return
	}

	// Fallback: return just the chart HTML for non-Datastar requests
	c.data = c.GenerateRandomData()
	chartComponent := TargetBarChart(c.data)

	w.Header().Set("Content-Type", "text/html")
	chartComponent.Render(r.Context(), w)
}

// updateHandler handles requests to increment/decrement a bar's completed count
func (c *Component) updateHandler(w http.ResponseWriter, r *http.Request) {
	// Expect datastar request
	if r.URL.Query().Get("datastar") == "" {
		http.Error(w, "datastar query param required", http.StatusBadRequest)
		return
	}

	// Read bar index and action from query parameters
	barIndexStr := r.URL.Query().Get("barIndex")
	action := r.URL.Query().Get("action")

	if barIndexStr == "" {
		http.Error(w, "missing barIndex parameter", http.StatusBadRequest)
		return
	}
	if action == "" {
		http.Error(w, "missing action parameter", http.StatusBadRequest)
		return
	}

	barIndex, err := strconv.Atoi(barIndexStr)
	if err != nil {
		http.Error(w, "invalid barIndex", http.StatusBadRequest)
		return
	}

	// Determine delta
	delta := 0
	switch action {
	case "increment":
		delta = 1
	case "decrement":
		delta = -1
	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	// Update bar
	if !c.updateBar(barIndex, delta) {
		http.Error(w, "bar index out of range", http.StatusBadRequest)
		return
	}

	// Create the target bar chart component with updated data
	chartComponent := TargetBarChart(c.data)

	// Create SSE response
	sse := datastar.NewSSE(w, r)

	// Render component to string
	var buf strings.Builder
	if err := chartComponent.Render(r.Context(), &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Patch the chart into the DOM
	sse.PatchElements(buf.String())
}

// signalsHandler handles requests for target bar chart signals
func (c *Component) signalsHandler(w http.ResponseWriter, r *http.Request) {
	// Example endpoint to patch signals
	sse := datastar.NewSSE(w, r)

	// Send current stats
	signals := map[string]interface{}{
		"lastUpdated": time.Now().Format(time.RFC3339),
		"totalBars":   len(c.data.Bars),
	}

	jsonData, _ := json.Marshal(signals)
	sse.PatchSignals(jsonData)
}

// RegisterRoutes registers HTTP routes for the target bar chart component
func (c *Component) RegisterRoutes(r chi.Router) {
	r.Get("/randomize", c.randomizeHandler)
	r.Get("/update", c.updateHandler)
	r.Get("/signals", c.signalsHandler)
}

// RegisterStatic registers static asset routes for the target bar chart component
func (c *Component) RegisterStatic(r chi.Router) {
	// Target bar chart doesn't have static assets yet
	// r.Handle("/assets/*", http.FileServer(http.Dir("./components/targetbarchart/assets")))
}
