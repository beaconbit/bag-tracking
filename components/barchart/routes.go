package barchart

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

var barColors = []string{
	"#4f46e5", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6",
	"#06b6d4", "#84cc16", "#f97316", "#ec4899", "#14b8a6",
}

var barLabels = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

// Component implements the bar chart component
type Component struct {
	rand *rand.Rand
}

// New creates a new bar chart component instance
func New() *Component {
	return &Component{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateRandomData creates random bar chart data
func (c *Component) GenerateRandomData() BarChartData {
	// Generate 5-8 random bars
	numBars := c.rand.Intn(4) + 5

	bars := make([]BarData, numBars)

	// Generate random values
	for i := 0; i < numBars; i++ {
		value := float64(c.rand.Intn(100) + 10)
		bars[i] = BarData{
			Label: barLabels[i%len(barLabels)],
			Color: barColors[i%len(barColors)],
			Value: value,
		}
	}

	data := BarChartData{
		ID:     "bar-chart",
		Title:  "Monthly Performance",
		Bars:   bars,
		Width:  600,
		Height: 400,
	}

	// Compute SVG and HTML
	data.SVG = data.GenerateSVGString()
	data.HTML = data.GenerateHTML()
	return data
}

// randomizeHandler handles requests to randomize the bar chart
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
		data := c.GenerateRandomData()

		// Create the bar chart component
		chartComponent := BarChart(data)

		// Create SSE response
		sse := datastar.NewSSE(w, r)

		// Render component to string
		var buf strings.Builder
		if err := chartComponent.Render(r.Context(), &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Patch the chart into the DOM (morphing into element with id="bar-chart")
		sse.PatchElements(buf.String())

		return
	}

	// Fallback: return just the chart HTML for non-Datastar requests
	data := c.GenerateRandomData()
	chartComponent := BarChart(data)

	w.Header().Set("Content-Type", "text/html")
	chartComponent.Render(r.Context(), w)
}

// signalsHandler handles requests for bar chart signals
func (c *Component) signalsHandler(w http.ResponseWriter, r *http.Request) {
	// Example endpoint to patch signals
	sse := datastar.NewSSE(w, r)

	// Send some example signals
	signals := map[string]interface{}{
		"lastUpdated": time.Now().Format(time.RFC3339),
		"totalBars":   7,
	}

	jsonData, _ := json.Marshal(signals)
	sse.PatchSignals(jsonData)
}

// RegisterRoutes registers HTTP routes for the bar chart component
func (c *Component) RegisterRoutes(r chi.Router) {
	r.Get("/randomize", c.randomizeHandler)
	r.Get("/signals", c.signalsHandler)
}

// RegisterStatic registers static asset routes for the bar chart component
func (c *Component) RegisterStatic(r chi.Router) {
	// Bar chart doesn't have static assets yet
	// r.Handle("/assets/*", http.FileServer(http.Dir("./components/barchart/assets")))
}
