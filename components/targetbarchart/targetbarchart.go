package targetbarchart

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

// BarData represents a single bar in the target chart
type BarData struct {
	Label     string // Product label
	Target    int    // Total number of sections (target)
	Completed int    // Number of completed sections (green)
	Color     string // Base color for incomplete sections (optional)
}

// TargetBarChartData represents a target bar chart component's data
type TargetBarChartData struct {
	ID         string
	Title      string
	XAxisLabel string
	YAxisLabel string
	Bars       []BarData
	Width      int
	Height     int
	BarHeight  int    // Fixed height for each bar
	SVG        string // Full SVG string
	HTML       string // Full chart HTML
}

// DefaultTargetBarChart creates a target bar chart with default values
func DefaultTargetBarChart() TargetBarChartData {
	return TargetBarChartData{
		ID:         "target-bar-chart",
		Title:      "Target Progress Chart",
		XAxisLabel: "Products",
		YAxisLabel: "Sections",
		Width:      800,
		Height:     500,
		BarHeight:  200, // Fixed height for each bar
		Bars:       []BarData{},
	}
}

// GenerateSVGString generates the SVG for a target bar chart
func (b TargetBarChartData) GenerateSVGString() string {
	component := TargetBarChartSVG(b)
	if html, err := renderComponentToString(component); err == nil {
		return html
	}
	return ""
}

// GenerateHTML generates the HTML for a target bar chart
func (b TargetBarChartData) GenerateHTML() string {
	component := TargetBarChartComponent(b)
	html, err := renderComponentToString(component)
	if err != nil {
		return ""
	}
	return html
}

// WithBars creates a copy with specified bars
func (b TargetBarChartData) WithBars(bars []BarData) TargetBarChartData {
	b.Bars = bars
	return b
}

// WithTitle creates a copy with specified title
func (b TargetBarChartData) WithTitle(title string) TargetBarChartData {
	b.Title = title
	return b
}

// WithAxisLabels creates a copy with specified axis labels
func (b TargetBarChartData) WithAxisLabels(xLabel, yLabel string) TargetBarChartData {
	b.XAxisLabel = xLabel
	b.YAxisLabel = yLabel
	return b
}
