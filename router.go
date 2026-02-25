package main

import (
	"github.com/go-chi/chi/v5"
	"bag-tracker/components/form"
)

// setupRouter configures the application router with all components
func setupRouter(r chi.Router) {
	// Mount component routers
	// Each component gets its own subrouter under /api/{component-name}

	// Form component
	formComp := form.New()
	formRouter := chi.NewRouter()
	formComp.RegisterRoutes(formRouter)
	r.Mount("/api/form", formRouter)

	// Form static assets
	formComp.RegisterStatic(r)
}
