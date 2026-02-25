package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"piechart-demo/components/barchart"
	"piechart-demo/components/piechart"
	"piechart-demo/components/stackedbarchart"
	"piechart-demo/components/targetbarchart"
	"piechart-demo/templates"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// homeHandler serves the component library home page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.Home()
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// piechartPageHandler serves a page with just the pie chart component
func piechartPageHandler(w http.ResponseWriter, r *http.Request) {
	piechartComp := piechart.New()
	data := piechartComp.GenerateRandomData()

	component := templates.PieChartPage(data.HTML)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// barchartPageHandler serves a page with just the bar chart component
func barchartPageHandler(w http.ResponseWriter, r *http.Request) {
	barchartComp := barchart.New()
	data := barchartComp.GenerateRandomData()

	component := templates.BarChartPage(data.HTML)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// buttonPageHandler serves a page with just the button component
func buttonPageHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.ButtonPage()
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// formPageHandler serves a page with just the form component
func formPageHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.FormPage()
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// targetbarchartPageHandler serves a page with just the target bar chart component
func targetbarchartPageHandler(w http.ResponseWriter, r *http.Request) {
	targetbarchartComp := targetbarchart.New()
	data := targetbarchartComp.GenerateRandomData()

	component := templates.TargetBarChartPage(data.HTML)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// stackedbarchartPageHandler serves a page with just the stacked bar chart component
func stackedbarchartPageHandler(w http.ResponseWriter, r *http.Request) {
	stackedbarchartComp := stackedbarchart.New()
	data := stackedbarchartComp.GenerateEmptyData()

	component := templates.StackedBarChartPage(data.HTML)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// piechartDemoHandler serves the original pie chart demo page (for backward compatibility)
func piechartDemoHandler(w http.ResponseWriter, r *http.Request) {
	piechartComp := piechart.New()
	data := piechartComp.GenerateRandomData()

	component := templates.Page(data)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// Middleware to log request duration
func requestTimer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("%s %s completed in %v", r.Method, r.URL.Path, duration)
	})
}

func main() {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(requestTimer)

	// Setup component API routes
	setupRouter(r)

	// Home page with component library
	r.Get("/", homeHandler)

	// Component pages (standalone views)
	r.Get("/component/piechart", piechartPageHandler)
	r.Get("/component/barchart", barchartPageHandler)
	r.Get("/component/button", buttonPageHandler)
	r.Get("/component/form", formPageHandler)
	r.Get("/component/targetbarchart", targetbarchartPageHandler)
	r.Get("/component/stackedbarchart", stackedbarchartPageHandler)

	// Original pie chart demo (backward compatibility)
	r.Get("/demo", piechartDemoHandler)

	// Serve global static files
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	fmt.Println("Press Ctrl+C to stop")

	log.Fatal(http.ListenAndServe(port, r))
}
