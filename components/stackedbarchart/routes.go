package stackedbarchart

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

// Component implements the stacked bar chart component
type Component struct {
	mu   sync.RWMutex
	data StackedBarChartData
}

// New creates a new stacked bar chart component instance
func New() *Component {
	comp := &Component{}
	comp.data = comp.GenerateEmptyData()
	return comp
}

func (c *Component) GenerateEmptyData() StackedBarChartData {
	data := DefaultStackedBarChart()
	data.SVG = data.GenerateSVGString()
	data.HTML = data.GenerateHTML()
	return data
}

// minuteIncrementHandler handles requests to add 1 minute delay to a machine
func (c *Component) minuteIncrementHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("minuteIncrementHandler: request received: %s %s", r.Method, r.URL.String())
	// Expect datastar request
	if r.URL.Query().Get("datastar") == "" {
		log.Printf("minuteIncrementHandler: missing datastar query param")
		http.Error(w, "datastar query param required", http.StatusBadRequest)
		return
	}

	// Read machine ID from query parameters
	machineIDStr := r.URL.Query().Get("machineId")
	if machineIDStr == "" {
		log.Printf("minuteIncrementHandler: missing machineId parameter")
		http.Error(w, "missing machineId parameter", http.StatusBadRequest)
		return
	}

	machineID, err := strconv.Atoi(machineIDStr)
	if err != nil || machineID < 0 || machineID >= 3 {
		log.Printf("minuteIncrementHandler: invalid machineId: %s", machineIDStr)
		http.Error(w, "invalid machineId", http.StatusBadRequest)
		return
	}

	log.Printf("minuteIncrementHandler: adding 1 minute delay to machine %d", machineID)
	// Add random delay to the specified machine
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.AddMinuteDelay(machineID)

	// Regenerate SVG and HTML
	log.Printf("minuteIncrementHandler: regenerating SVG and HTML")
	c.data.SVG = c.data.GenerateSVGString()
	c.data.HTML = c.data.GenerateHTML()

	// Create the stacked bar chart component with updated data
	chartComponent := StackedBarChart(c.data, true)

	// Create SSE response
	sse := datastar.NewSSE(w, r)

	// Render component to string
	var buf strings.Builder
	if err := chartComponent.Render(r.Context(), &buf); err != nil {
		log.Printf("minuteIncrementHandler: error rendering component: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := buf.String()
	log.Printf("minuteIncrementHandler: patching DOM with HTML length %d", len(html))
	// Patch the chart into the DOM
	sse.PatchElements(html)
}

// advanceHourHandler handles requests to advance the chart by one hour
func (c *Component) advanceHourHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("advanceHourHandler: request received: %s %s", r.Method, r.URL.String())
	// Expect datastar request
	if r.URL.Query().Get("datastar") == "" {
		log.Printf("advanceHourHandler: missing datastar query param")
		http.Error(w, "datastar query param required", http.StatusBadRequest)
		return
	}

	log.Printf("advanceHourHandler: advancing chart hour")
	// Advance chart by one minute
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.AdvanceHour()

	// Regenerate SVG and HTML
	log.Printf("advanceHourHandler: regenerating SVG and HTML")
	c.data.SVG = c.data.GenerateSVGString()
	c.data.HTML = c.data.GenerateHTML()

	// Create the stacked bar chart component with updated data
	chartComponent := StackedBarChart(c.data, true)

	// Create SSE response
	sse := datastar.NewSSE(w, r)

	// Render component to string
	var buf strings.Builder
	if err := chartComponent.Render(r.Context(), &buf); err != nil {
		log.Printf("advanceHourHandler: error rendering component: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := buf.String()
	log.Printf("advanceHourHandler: patching DOM with HTML length %d", len(html))
	// Patch the chart into the DOM
	sse.PatchElements(html)
}

// clockTickHandler handles requests to update the live clock via SSE long-polling
func (c *Component) clockTickHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("clockTickHandler: request received: %s %s", r.Method, r.URL.String())

	// If not a datastar request, return plain HTML clock (backward compatibility)
	if r.URL.Query().Get("datastar") == "" {
		log.Printf("clockTickHandler: non-datastar request, returning plain HTML clock")
		clockHTML := c.data.GenerateClockHTML()
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(clockHTML))
		return
	}

	log.Printf("clockTickHandler: datastar request detected, setting up SSE")
	// Set up SSE connection
	sse := datastar.NewSSE(w, r)
	// Get flusher if available
	flusher, _ := w.(http.Flusher)

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("clockTickHandler panic recovered: %v", r)
		}
	}()

	// Get initial time
	now := time.Now()
	lastSecond := now.Second()
	lastHour := now.Hour()
	log.Printf("clockTickHandler: initial time - hour:%d second:%d", lastHour, lastSecond)

	// Track chart's current hour (for detecting when to advance)
	c.mu.RLock()
	chartHour := c.data.CurrentTime.Hour()
	c.mu.RUnlock()
	log.Printf("clockTickHandler: chart hour: %d", chartHour)

	// Create a ticker that fires every 100ms for time checking
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	log.Printf("clockTickHandler: entering SSE loop")
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			log.Printf("clockTickHandler: client disconnected")
			return
		case <-ticker.C:
			now := time.Now()
			currentSecond := now.Second()
			currentHour := now.Hour()

			// Check if wall-clock hour has changed
			if currentHour != lastHour {
				log.Printf("clockTickHandler: hour changed from %d to %d", lastHour, currentHour)
				lastHour = currentHour

				// Check if chart's hour is behind wall-clock hour
				c.mu.Lock()
				chartHour = c.data.CurrentTime.Hour()
				log.Printf("clockTickHandler: BEFORE ADVANCE - chart hour: %d, wall-clock hour: %d, data.CurrentTime: %s",
					chartHour, currentHour, c.data.CurrentTime.Format("15:04:05"))

				// Log hours array before advance
				log.Printf("clockTickHandler: BEFORE ADVANCE - Hours array length: %d", len(c.data.Hours))
				for i, hour := range c.data.Hours {
					log.Printf("clockTickHandler: BEFORE hour[%d] - offset: %d, timestamp: %s, totalDelay: %d",
						i, hour.HourOffset, hour.Timestamp.Format("15:04:05"), hour.TotalDelay)
				}

				if chartHour != currentHour {
					log.Printf("clockTickHandler: chart hour (%d) behind wall-clock hour (%d), advancing", chartHour, currentHour)
					// BRANCH: Advance chart data and send entire modified page with new bar locations
					c.data.AdvanceHour()

					// Log hours array after advance
					log.Printf("clockTickHandler: AFTER ADVANCE - Hours array length: %d, data.CurrentTime: %s",
						len(c.data.Hours), c.data.CurrentTime.Format("15:04:05"))
					for i, hour := range c.data.Hours {
						log.Printf("clockTickHandler: AFTER hour[%d] - offset: %d, timestamp: %s, totalDelay: %d",
							i, hour.HourOffset, hour.Timestamp.Format("15:04:05"), hour.TotalDelay)
					}

					c.data.SVG = c.data.GenerateSVGString()
					c.data.HTML = c.data.GenerateHTML()
					chartComponent := StackedBarChart(c.data, true)
					var buf strings.Builder
					if err := chartComponent.Render(r.Context(), &buf); err != nil {
						c.mu.Unlock()
						log.Printf("clockTickHandler: error rendering chart component: %v", err)
						// If rendering fails, abort the connection
						return
					}
					html := buf.String()
					c.mu.Unlock()
					sse.PatchElements(html)
					if flusher != nil {
						flusher.Flush()
					}
					log.Printf("clockTickHandler: sent updated chart HTML")
					// After advancing the chart, we've already sent a fresh clock;
					// no need to send a separate clock tick update this iteration.
					continue
				} else {
					log.Printf("clockTickHandler: chart hour (%d) already up to date with wall-clock hour (%d), no advance needed", chartHour, currentHour)
				}
				c.mu.Unlock()
			}

			// Check if wall-clock second has changed
			if currentSecond != lastSecond {
				log.Printf("clockTickHandler: second changed from %d to %d", lastSecond, currentSecond)
				lastSecond = currentSecond
				// BRANCH: Simple clock tick update – send only the new clock HTML
				c.mu.RLock()
				clockHTML := c.data.GenerateClockHTML()
				c.mu.RUnlock()
				log.Printf("clockTickHandler: clock HTML length %d", len(clockHTML))
				sse.PatchElements(clockHTML)
				if flusher != nil {
					flusher.Flush()
				}
				log.Printf("clockTickHandler: sent clock tick update")
			}
		}
	}
}

// signalsHandler handles requests for stacked bar chart signals
func (c *Component) signalsHandler(w http.ResponseWriter, r *http.Request) {
	// Example endpoint to patch signals
	sse := datastar.NewSSE(w, r)

	c.mu.RLock()
	defer c.mu.RUnlock()
	// Send current stats
	signals := map[string]interface{}{
		"lastUpdated": time.Now().Format(time.RFC3339),
		"totalHours":  len(c.data.Hours),
		"currentTime": c.data.CurrentTime.Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(signals)
	sse.PatchSignals(jsonData)
}

// RegisterRoutes registers HTTP routes for the stacked bar chart component
func (c *Component) RegisterRoutes(r chi.Router) {
	r.Get("/minuteincrement", c.minuteIncrementHandler)
	r.Get("/advance", c.advanceHourHandler)
	r.Get("/tick", c.clockTickHandler)
	r.Get("/signals", c.signalsHandler)
}

// RegisterStatic registers static asset routes for the stacked bar chart component
func (c *Component) RegisterStatic(r chi.Router) {
	// Stacked bar chart doesn't have static assets yet
	// r.Handle("/assets/*", http.FileServer(http.Dir("./components/stackedbarchart/assets")))
}
