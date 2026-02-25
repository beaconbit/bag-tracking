package templates

import (
	"bytes"
	"context"
	"piechart-demo/components/piechart"
	"strings"
	"testing"
)

func TestPageButton(t *testing.T) {
	colors := []string{"#4f46e5", "#10b981"}
	labels := []string{"Tech", "Health"}
	sectors := []piechart.Sector{
		{Label: labels[0], Color: colors[0], Value: 50, Percentage: 50},
		{Label: labels[1], Color: colors[1], Value: 50, Percentage: 50},
	}
	data := piechart.PieChartData{
		ID:      "pie-chart",
		Title:   "Test",
		Sectors: sectors,
		Width:   500,
		Height:  500,
	}
	data.RenderData = piechart.ComputeRenderData(data)
	data.SVG = piechart.GenerateSVGString(data)
	data.HTML = piechart.GenerateChartHTML(data)

	var buf bytes.Buffer
	component := Page(data)
	component.Render(context.Background(), &buf)

	html := buf.String()
	if !strings.Contains(html, `data-on:click="@get('/api/piechart/randomize')"`) {
		t.Errorf("Button attribute incorrect. HTML snippet:\n%s", html)
	}
}
