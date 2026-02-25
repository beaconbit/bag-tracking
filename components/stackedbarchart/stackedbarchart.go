package stackedbarchart

import (
	"context"
	"log"
	"strings"
	"time"

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

// Machine represents a washing machine with its delay data
type Machine struct {
	ID         int
	Name       string
	Color      string
	TotalDelay int // Running total delay in seconds
}

// HourBar represents a single hour's stacked bar data
type HourBar struct {
	HourOffset    int // 0 = current hour, -1 = 1 hr ago, ..., -9 = 9 hr ago
	Timestamp     time.Time
	MachineDelays [3]int // Delays for each machine in seconds
	TotalDelay    int    // Sum of all machine delays for this hour
}

// StackedBarChartData represents the stacked bar chart component's data
type StackedBarChartData struct {
	ID          string
	Title       string
	XAxisLabel  string
	YAxisLabel  string
	Hours       []HourBar // 10 hours, index 0 is oldest (-9), index 9 is current (0)
	Machines    [3]Machine
	CurrentTime time.Time
	Width       int
	Height      int
	BarWidth    int // Width of each bar
	SVG         string
	HTML        string
	ClockHTML   string // Live clock display
	LegendHTML  string // Dynamic legend with buttons
}

// DefaultStackedBarChart creates a stacked bar chart with default values
func DefaultStackedBarChart() StackedBarChartData {
	now := time.Now()

	// Initialize machines
	machines := [3]Machine{
		{
			ID:         0,
			Name:       "Continuous Batch Washer 1",
			Color:      "#d8b4fe", // Light Purple
			TotalDelay: 0,
		},
		{
			ID:         1,
			Name:       "Continuous Batch Washer 2",
			Color:      "#a855f7", // Medium Purple
			TotalDelay: 0,
		},
		{
			ID:         2,
			Name:       "Continuous Batch Washer 3",
			Color:      "#7c3aed", // Dark Purple
			TotalDelay: 0,
		},
	}

	// Initialize 10 hours of data (9 hours ago to current hour)
	hours := make([]HourBar, 10)
	for i := 0; i < 10; i++ {
		offset := i - 9 // -9 (oldest) to 0 (current)
		timestamp := now.Add(time.Duration(offset) * time.Hour)
		hours[i] = HourBar{
			HourOffset:    offset,
			Timestamp:     timestamp,
			MachineDelays: [3]int{0, 0, 0},
			TotalDelay:    0,
		}
	}

	return StackedBarChartData{
		ID:          "stacked-bar-chart",
		Title:       "Washing Machine Delay Monitor",
		XAxisLabel:  "Hours Ago",
		YAxisLabel:  "Delay (minutes)",
		Hours:       hours,
		Machines:    machines,
		CurrentTime: now,
		Width:       1000,
		Height:      600,
		BarWidth:    80,
	}
}

// GenerateClockHTML generates the live clock display
func (s *StackedBarChartData) GenerateClockHTML() string {
	now := time.Now()
	currentTime := now.Format("15:04:05")
	clockID := s.ID + "-clock"
	component := Clock(currentTime, true, clockID)
	if html, err := renderComponentToString(component); err == nil {
		return html
	}
	return ""
}

// GenerateLegendHTML generates the dynamic legend with buttons and running counters
func (s *StackedBarChartData) GenerateLegendHTML() string {
	legendID := s.ID + "-legend"
	component := Legend(s.Machines, legendID, s.ID)
	if html, err := renderComponentToString(component); err == nil {
		return html
	}
	return ""
}

// GenerateSVGString generates the SVG for the stacked bar chart
func (s *StackedBarChartData) GenerateSVGString() string {
	component := StackedBarChartSVG(*s)
	if html, err := renderComponentToString(component); err == nil {
		return html
	}
	return ""
}

// GenerateHTML generates the full HTML for the stacked bar chart component
func (s *StackedBarChartData) GenerateHTML() string {
	log.Printf("GenerateHTML called, rendering with autoStart=true")
	// Update component fields
	now := time.Now()
	currentTime := now.Format("15:04:05")
	clockID := s.ID + "-clock"
	clockComponent := Clock(currentTime, true, clockID)
	clockHTML, err := renderComponentToString(clockComponent)
	if err != nil {
		clockHTML = ""
	}
	s.ClockHTML = clockHTML
	s.LegendHTML = s.GenerateLegendHTML()
	s.SVG = s.GenerateSVGString()

	// Render full component using templ
	component := StackedBarChartComponent(*s, true)
	if html, err := renderComponentToString(component); err == nil {
		log.Printf("GenerateHTML: rendered successfully, length=%d", len(html))
		return html
	}
	log.Printf("GenerateHTML: rendering failed")
	return ""
}

// AddMinuteDelay adds a 1 minute delay (60 seconds) to a machine's current hour
func (s *StackedBarChartData) AddMinuteDelay(machineID int) {
	if machineID < 0 || machineID >= len(s.Machines) {
		return
	}

	// Add 1 minute delay (60 seconds)
	delay := 60

	// Update current hour (last in array)
	currentHour := &s.Hours[len(s.Hours)-1]
	currentHour.MachineDelays[machineID] += delay
	currentHour.TotalDelay += delay

	// Update machine totals
	s.Machines[machineID].TotalDelay += delay
}

// AdvanceHour shifts the chart by one hour, dropping oldest, adding new current hour
func (s *StackedBarChartData) AdvanceHour() {
	log.Printf("AdvanceHour: START - CurrentTime: %s, Hours length: %d", s.CurrentTime.Format("15:04:05"), len(s.Hours))

	// Log current hour offsets and totals
	for i, hour := range s.Hours {
		log.Printf("AdvanceHour: BEFORE hour[%d] - offset: %d, timestamp: %s, totalDelay: %d, machineDelays: %v",
			i, hour.HourOffset, hour.Timestamp.Format("15:04:05"), hour.TotalDelay, hour.MachineDelays)
	}

	// Log machine current delays
	for i, machine := range s.Machines {
		log.Printf("AdvanceHour: BEFORE machine[%d] - ID: %d, TotalDelay: %d",
			i, machine.ID, machine.TotalDelay)
	}

	// Drop oldest hour (index 0) and shift remaining left
	// Shift slice left by one (move elements 1..9 to 0..8)
	for i := 0; i < 9; i++ {
		s.Hours[i] = s.Hours[i+1]
		// Update offset: each hour becomes one hour older relative to new current time
		s.Hours[i].HourOffset -= 1
	}

	// Create new current hour with zero delays
	now := time.Now()
	newCurrentHour := HourBar{
		HourOffset:    0,
		Timestamp:     now,
		MachineDelays: [3]int{0, 0, 0},
		TotalDelay:    0,
	}

	// Replace last element with new current hour
	s.Hours[9] = newCurrentHour

	// Update current time
	s.CurrentTime = now

	// Reset machine TotalDelay fields for new hour
	for i := range s.Machines {
		s.Machines[i].TotalDelay = 0
	}

	// Log after state
	log.Printf("AdvanceHour: AFTER - CurrentTime: %s, Hours length: %d", s.CurrentTime.Format("15:04:05"), len(s.Hours))
	for i, hour := range s.Hours {
		log.Printf("AdvanceHour: AFTER hour[%d] - offset: %d, timestamp: %s, totalDelay: %d, machineDelays: %v",
			i, hour.HourOffset, hour.Timestamp.Format("15:04:05"), hour.TotalDelay, hour.MachineDelays)
	}

	// Log machine current delays after reset
	for i, machine := range s.Machines {
		log.Printf("AdvanceHour: AFTER machine[%d] - ID: %d, TotalDelay: %d",
			i, machine.ID, machine.TotalDelay)
	}

	log.Printf("AdvanceHour: COMPLETE")
}
