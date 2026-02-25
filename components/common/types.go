package common

import "github.com/go-chi/chi/v5"

// Component defines the interface for reusable components
type Component interface {
	// RegisterRoutes registers HTTP routes for the component under the given router
	RegisterRoutes(r chi.Router)
	// RegisterStatic registers static asset routes for the component
	RegisterStatic(r chi.Router)
}
