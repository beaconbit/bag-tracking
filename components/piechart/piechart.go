package piechart

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
)

type Point struct {
	X float64
	Y float64
}

type Sector struct {
	Label      string
	Color      string
	Value      float64
	Percentage float64
}

type PieChartData struct {
	ID         string
	Title      string
	Sectors    []Sector
	Width      int
	Height     int
	RenderData []SectorRenderData // Precomputed for template
	SVG        string             // Full SVG string
	HTML       string             // Full chart HTML
}

func renderComponentToString(c templ.Component) (string, error) {
	var buf strings.Builder
	ctx := context.Background()
	if err := c.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetSectorPath(index int, sector Sector, sectors []Sector) string {
	// Calculate start and end angles based on cumulative percentages
	total := 0.0
	for i := 0; i < len(sectors); i++ {
		if i == index {
			break
		}
		total += sectors[i].Percentage
	}

	startAngle := total * 3.6 // Convert percentage to degrees (0-360)
	endAngle := startAngle + sector.Percentage*3.6

	// Convert to radians
	startRad := startAngle * math.Pi / 180
	endRad := endAngle * math.Pi / 180

	// Center and radius
	cx, cy, r := 200.0, 200.0, 150.0

	// Calculate points
	x1 := cx + r*math.Cos(startRad)
	y1 := cy + r*math.Sin(startRad)
	x2 := cx + r*math.Cos(endRad)
	y2 := cy + r*math.Sin(endRad)

	// Large arc flag: 1 if angle > 180 degrees
	largeArcFlag := 0
	if sector.Percentage > 50 {
		largeArcFlag = 1
	}

	// Create SVG path
	return fmt.Sprintf("M%.2f,%.2f L%.2f,%.2f A%.2f,%.2f 0 %d 1 %.2f,%.2f Z",
		cx, cy, x1, y1, r, r, largeArcFlag, x2, y2)
}

func GetLabelPosition(index int, sector Sector, sectors []Sector) Point {
	// Calculate middle angle of sector
	total := 0.0
	for i := 0; i < len(sectors); i++ {
		if i == index {
			break
		}
		total += sectors[i].Percentage
	}

	middleAngle := (total + sector.Percentage/2) * 3.6
	middleRad := middleAngle * math.Pi / 180

	// Position at 70% of radius
	cx, cy, r := 200.0, 200.0, 105.0

	return Point{
		X: cx + r*math.Cos(middleRad),
		Y: cy + r*math.Sin(middleRad),
	}
}

type SectorRenderData struct {
	Path           string
	LabelX         string
	LabelY         string
	Label          string
	Color          string
	Percentage     float64
	PercentageText string
}

func ComputeRenderData(data PieChartData) []SectorRenderData {
	result := make([]SectorRenderData, len(data.Sectors))
	for i, sector := range data.Sectors {
		pos := GetLabelPosition(i, sector, data.Sectors)
		result[i] = SectorRenderData{
			Path:           GetSectorPath(i, sector, data.Sectors),
			LabelX:         fmt.Sprintf("%.1f", pos.X),
			LabelY:         fmt.Sprintf("%.1f", pos.Y),
			Label:          sector.Label,
			Color:          sector.Color,
			Percentage:     sector.Percentage,
			PercentageText: FormatPercentage(sector.Percentage),
		}
	}
	return result
}

func GenerateSVGString(data PieChartData) string {
	component := PieChartSVG(data)
	if html, err := renderComponentToString(component); err == nil {
		return html
	}
	return ""
}

func GenerateChartHTML(data PieChartData) string {
	component := PieChartContainer(data)
	if html, err := renderComponentToString(component); err == nil {
		return html
	}
	return ""
}

func FormatPercentage(p float64) string {
	return fmt.Sprintf("%.1f%%", p)
}
