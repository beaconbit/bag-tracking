package main

import (
	"github.com/go-chi/chi/v5"
	"piechart-demo/components/barchart"
	"piechart-demo/components/button"
	"piechart-demo/components/form"
	"piechart-demo/components/piechart"
	"piechart-demo/components/stackedbarchart"
	"piechart-demo/components/targetbarchart"
)

// setupRouter configures the application router with all components
func setupRouter(r chi.Router) {
	// Mount component routers
	// Each component gets its own subrouter under /api/{component-name}

	// Piechart component
	piechartComp := piechart.New()
	piechartRouter := chi.NewRouter()
	piechartComp.RegisterRoutes(piechartRouter)
	r.Mount("/api/piechart", piechartRouter)

	// Bar chart component
	barchartComp := barchart.New()
	barchartRouter := chi.NewRouter()
	barchartComp.RegisterRoutes(barchartRouter)
	r.Mount("/api/barchart", barchartRouter)

	// Button component
	buttonComp := button.New()
	buttonRouter := chi.NewRouter()
	buttonComp.RegisterRoutes(buttonRouter)
	r.Mount("/api/button", buttonRouter)

	// Form component
	formComp := form.New()
	formRouter := chi.NewRouter()
	formComp.RegisterRoutes(formRouter)
	r.Mount("/api/form", formRouter)

	// Target bar chart component
	targetbarchartComp := targetbarchart.New()
	targetbarchartRouter := chi.NewRouter()
	targetbarchartComp.RegisterRoutes(targetbarchartRouter)
	r.Mount("/api/targetbarchart", targetbarchartRouter)

	// Stacked bar chart component
	stackedbarchartComp := stackedbarchart.New()
	stackedbarchartRouter := chi.NewRouter()
	stackedbarchartComp.RegisterRoutes(stackedbarchartRouter)
	r.Mount("/api/stackedbarchart", stackedbarchartRouter)

	// Mount component static assets
	// Each component's static assets are under /static/{component-name}

	// Piechart static assets
	piechartComp.RegisterStatic(r)

	// Bar chart static assets
	barchartComp.RegisterStatic(r)

	// Button static assets
	buttonComp.RegisterStatic(r)

	// Form static assets
	formComp.RegisterStatic(r)

	// Target bar chart static assets
	targetbarchartComp.RegisterStatic(r)

	// Stacked bar chart static assets
	stackedbarchartComp.RegisterStatic(r)
}
