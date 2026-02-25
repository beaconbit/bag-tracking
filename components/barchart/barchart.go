package barchart

import (
	"context"
	"strings"

	"github.com/a-h/templ"
)

// BarData represents a single bar in the chart
type BarData struct {
	Label string
	Value float64
	Color string
}

// BarChartData represents a bar chart component's data
type BarChartData struct {
	ID     string
	Title  string
	Bars   []BarData
	Width  int
	Height int
	SVG    string // Full SVG string
	HTML   string // Full chart HTML
}

func renderComponentToString(c templ.Component) (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := c.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// DefaultBarChart creates a bar chart with default values
func DefaultBarChart() BarChartData {
	return BarChartData{
		ID:     "bar-chart",
		Title:  "Bar Chart Demo",
		Width:  600,
		Height: 400,
		Bars:   []BarData{},
	}
}

// GenerateSVGString generates the SVG for a bar chart
func (b BarChartData) GenerateSVGString() string {
	component := BarChartSVG(b)
	if html, err := renderComponentToString(component); err == nil {
		return html
	}
	return ""
}

// GenerateHTML generates the HTML for a bar chart
func (b BarChartData) GenerateHTML() string {
	component := BarChartContainer(b)
	if html, err := renderComponentToString(component); err == nil {
		return html
	}
	return ""
}

// WithBars creates a copy with specified bars
func (b BarChartData) WithBars(bars []BarData) BarChartData {
	b.Bars = bars
	return b
}

// WithTitle creates a copy with specified title
func (b BarChartData) WithTitle(title string) BarChartData {
	b.Title = title
	return b
}
