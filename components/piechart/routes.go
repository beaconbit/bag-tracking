package piechart

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

var colors = []string{
	"#4f46e5", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6",
	"#06b6d4", "#84cc16", "#f97316", "#ec4899", "#14b8a6",
}

var labels = []string{
	"Technology", "Healthcare", "Finance", "Energy", "Consumer",
	"Industrials", "Materials", "Real Estate", "Utilities", "Communications",
}

// Component implements the piechart component
type Component struct {
	rand *rand.Rand
}

// New creates a new piechart component instance
func New() *Component {
	return &Component{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Name returns the component name
func (c *Component) Name() string {
	return "piechart"
}

// GenerateRandomData creates random pie chart data
func (c *Component) GenerateRandomData() PieChartData {
	// Generate 5-8 random sectors
	numSectors := c.rand.Intn(4) + 5

	sectors := make([]Sector, numSectors)
	totalValue := 0.0

	// Generate random values
	for i := 0; i < numSectors; i++ {
		value := float64(c.rand.Intn(100) + 10)
		sectors[i] = Sector{
			Label: labels[i%len(labels)],
			Color: colors[i%len(colors)],
			Value: value,
		}
		totalValue += value
	}

	// Calculate percentages
	for i := range sectors {
		sectors[i].Percentage = (sectors[i].Value / totalValue) * 100
	}

	data := PieChartData{
		ID:      "pie-chart",
		Title:   "Market Distribution",
		Sectors: sectors,
		Width:   500,
		Height:  500,
	}

	// Compute render data, SVG and full HTML
	data.RenderData = ComputeRenderData(data)
	data.SVG = GenerateSVGString(data)
	data.HTML = GenerateChartHTML(data)
	return data
}

// randomizeHandler handles requests to randomize the pie chart
func (c *Component) randomizeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if this is a Datastar request by looking for the datastar query param
	if r.URL.Query().Get("datastar") != "" {
		// Read signals from request (though we don't need them for this demo)
		var signals map[string]interface{}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate new random data
		data := c.GenerateRandomData()

		// Create the pie chart component
		chartComponent := PieChart(data)

		// Create SSE response
		sse := datastar.NewSSE(w, r)

		// Render component to string
		var buf strings.Builder
		if err := chartComponent.Render(r.Context(), &buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Patch the chart into the DOM (morphing into element with id="pie-chart")
		sse.PatchElements(buf.String())

		return
	}

	// Fallback: return just the chart HTML for non-Datastar requests
	data := c.GenerateRandomData()
	chartComponent := PieChart(data)

	w.Header().Set("Content-Type", "text/html")
	chartComponent.Render(r.Context(), w)
}

// signalsHandler handles requests for pie chart signals
func (c *Component) signalsHandler(w http.ResponseWriter, r *http.Request) {
	// Example endpoint to patch signals
	sse := datastar.NewSSE(w, r)

	// Send some example signals
	signals := map[string]interface{}{
		"lastUpdated":  time.Now().Format(time.RFC3339),
		"totalSectors": 7,
	}

	jsonData, _ := json.Marshal(signals)
	sse.PatchSignals(jsonData)
}

// RegisterRoutes registers HTTP routes for the piechart component
func (c *Component) RegisterRoutes(r chi.Router) {
	r.Get("/randomize", c.randomizeHandler)
	r.Get("/signals", c.signalsHandler)
}

// RegisterStatic registers static asset routes for the piechart component
func (c *Component) RegisterStatic(r chi.Router) {
	// Piechart doesn't have static assets yet
	// r.Handle("/assets/*", http.FileServer(http.Dir("./components/piechart/assets")))
}
