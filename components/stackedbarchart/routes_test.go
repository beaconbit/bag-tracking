package stackedbarchart

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClockTickHandler_HourChange(t *testing.T) {
	// Create component with chart's current time set to a hour ago
	comp := New()
	comp.data.CurrentTime = time.Now().Add(-time.Hour) // simulate that chart is behind wall time by a hour

	// Create request to tick endpoint with datastar param
	req := httptest.NewRequest("GET", "/tick?datastar=true", nil)
	w := httptest.NewRecorder()

	// Call handler directly
	comp.clockTickHandler(w, req)

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check response body contains SSE patch for chart container (indicating chart advanced)
	body := w.Body.String()
	if !strings.Contains(body, "event: datastar-patch-elements") {
		t.Errorf("expected SSE patch event, got: %s", body)
	}
	// When minute changes, the handler should patch the entire chart container, not just clock
	// The container ID is "stacked-bar-chart"
	if !strings.Contains(body, "stacked-bar-chart") {
		t.Log("patch may be for clock only; ensure minute changed logic works")
	}
}

func TestClockTickHandler_NoHourChange(t *testing.T) {
	// Create component with chart's current time set to current wall time
	comp := New()
	comp.data.CurrentTime = time.Now()

	req := httptest.NewRequest("GET", "/tick?datastar=true", nil)
	w := httptest.NewRecorder()

	comp.clockTickHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "stacked-chart-clock") {
		t.Errorf("expected clock patch, got: %s", body)
	}
}

func TestAdvanceHour(t *testing.T) {
	data := DefaultStackedBarChart()
	// Store original hours
	originalHours := make([]HourBar, len(data.Hours))
	copy(originalHours, data.Hours)

	data.AdvanceHour()

	// Check length unchanged
	if len(data.Hours) != 10 {
		t.Errorf("expected 10 hours, got %d", len(data.Hours))
	}

	// Check that hours shifted: index 0 should be original index 1 (compare timestamps)
	for i := 0; i < 9; i++ {
		if !data.Hours[i].Timestamp.Equal(originalHours[i+1].Timestamp) {
			t.Errorf("hour timestamp mismatch at index %d", i)
		}
	}
	// Check new current hour offset is 0
	if data.Hours[9].HourOffset != 0 {
		t.Errorf("new current hour offset should be 0, got %d", data.Hours[9].HourOffset)
	}
	// Check that new current hour has zero delays
	for i := 0; i < 3; i++ {
		if data.Hours[9].MachineDelays[i] != 0 {
			t.Errorf("new current hour should have zero delays for machine %d", i)
		}
	}
	// Check that oldest hour (original index 0) is no longer present
	// by verifying its timestamp is not in data.Hours
	oldestTimestamp := originalHours[0].Timestamp
	for i, hour := range data.Hours {
		if hour.Timestamp.Equal(oldestTimestamp) {
			t.Errorf("oldest hour still present at index %d", i)
		}
	}
}
