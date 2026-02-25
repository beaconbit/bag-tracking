package main

import (
	"bag-tracker/database"
	"bag-tracker/templates"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"strings"
	"time"
)

// homeHandler serves the component library home page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.Home()
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// formPageHandler serves a page with just the form component
func formPageHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.FormPage()
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// railPageHandler serves a rail management page
func railPageHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	// Validate rail type
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}
	component := templates.RailPage(railType)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// addBagHandler handles adding or updating a bag on a rail
func addBagHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	// Validate rail type
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}

	var request struct {
		BagNumber string `json:"bag_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := database.AddOrUpdateBag(railType, request.BagNumber); err != nil {
		log.Printf("Error adding/updating bag: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// listBagsHandler returns all bags for a rail
func listBagsHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	// Validate rail type
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}

	bags, err := database.GetBags(railType)
	if err != nil {
		log.Printf("Error getting bags: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bags)
}

// removeBagHandler handles removing a bag from a rail
func removeBagHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	// Validate rail type
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}

	var request struct {
		BagNumber string `json:"bag_number"`
		Anonymous bool   `json:"anonymous"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Handle anonymous bag removal
	if request.Anonymous {
		if err := database.RemoveAnonymousBag(railType); err != nil {
			log.Printf("Error creating anonymous bag: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	if err := database.RemoveBag(railType, request.BagNumber); err != nil {
		log.Printf("Error removing bag: %v", err)
		// Check if error is "bag not found"
		if strings.Contains(err.Error(), "bag not found") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// createWorkOrderHandler handles creating a new work order (request fix)
func createWorkOrderHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}

	var request struct {
		BagNumber int             `json:"bag_number"`
		Flags     map[string]bool `json:"flags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate bag number
	if request.BagNumber < 0 {
		http.Error(w, "Bag number must be non-negative", http.StatusBadRequest)
		return
	}

	// Ensure at least one flag is true (excluding the two auto-set flags)
	hasTrue := false
	for key, val := range request.Flags {
		if key == "work_request_order" || key == "work_completion_order" {
			continue
		}
		if val {
			hasTrue = true
			break
		}
	}
	if !hasTrue {
		http.Error(w, "At least one checkbox must be selected", http.StatusBadRequest)
		return
	}

	// Add the auto-set flags to the map
	if request.Flags == nil {
		request.Flags = make(map[string]bool)
	}
	request.Flags["work_request_order"] = true
	request.Flags["work_completion_order"] = false

	if err := database.CreateWorkOrder(railType, request.BagNumber, request.Flags); err != nil {
		log.Printf("Error creating work order: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If bag exists in bag-rail table, deactivate it (set fixed = false, in_production = false)
	if err := database.DeactivateBagIfExists(railType, request.BagNumber); err != nil {
		// Log error but don't fail the request - work order was created successfully
		log.Printf("Warning: failed to deactivate bag %d in rail %s: %v", request.BagNumber, railType, err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// listWorkOrdersHandler returns all work orders for a rail
func listWorkOrdersHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}

	orders, err := database.GetWorkOrders(railType)
	if err != nil {
		log.Printf("Error getting work orders: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// recordFixHandler handles batch recording of work completions
func recordFixHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}

	var request struct {
		Fixes []struct {
			BagNumber int             `json:"bag_number"`
			Flags     map[string]bool `json:"flags"`
		} `json:"fixes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate at least one fix
	if len(request.Fixes) == 0 {
		http.Error(w, "No fixes to record", http.StatusBadRequest)
		return
	}

	// Process each fix
	for _, fix := range request.Fixes {
		if fix.BagNumber < 0 {
			http.Error(w, "Bag number must be non-negative", http.StatusBadRequest)
			return
		}
		// Ensure flags map exists
		if fix.Flags == nil {
			fix.Flags = make(map[string]bool)
		}
		// Set completion flags
		fix.Flags["work_request_order"] = false
		fix.Flags["work_completion_order"] = true

		// Create work completion entry
		if err := database.CreateWorkOrder(railType, fix.BagNumber, fix.Flags); err != nil {
			log.Printf("Error recording fix for bag %d: %v", fix.BagNumber, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		// If bag exists in bag-rail table, mark it as fixed (set fixed = true, in_production = false)
		if err := database.MarkBagAsFixedIfExists(railType, fix.BagNumber); err != nil {
			// Log error but don't fail the request - work order was created successfully
			log.Printf("Warning: failed to mark bag %d as fixed in rail %s: %v", fix.BagNumber, railType, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// requestFixPageHandler serves the request fix form page
func requestFixPageHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}
	component := templates.RequestFixPage(railType)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// historyPageHandler serves the work order history page
func historyPageHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}
	component := templates.HistoryPage(railType)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// recordFixPageHandler serves the record fix page
func recordFixPageHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}
	component := templates.RecordFixPage(railType)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// addBagPageHandler serves the add bag page
func addBagPageHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}
	component := templates.AddBagPage(railType)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// removeBagPageHandler serves the remove bag page
func removeBagPageHandler(w http.ResponseWriter, r *http.Request) {
	railType := chi.URLParam(r, "railType")
	validTypes := map[string]bool{
		"sorting": true,
		"clean":   true,
		"ironer":  true,
	}
	if !validTypes[railType] {
		http.Error(w, "Invalid rail type", http.StatusNotFound)
		return
	}
	component := templates.RemoveBagPage(railType)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// healthHandler checks database connectivity
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := database.DB.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database unavailable"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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
	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

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

	r.Get("/component/form", formPageHandler)

	r.Get("/rail/{railType}", railPageHandler)
	r.Get("/rail/{railType}/add-bag", addBagPageHandler)
	r.Get("/rail/{railType}/remove-bag", removeBagPageHandler)
	r.Get("/rail/{railType}/request-fix", requestFixPageHandler)
	r.Get("/rail/{railType}/record-fix", recordFixPageHandler)
	r.Get("/rail/{railType}/history", historyPageHandler)

	r.Post("/api/{railType}/add-bag", addBagHandler)
	r.Post("/api/{railType}/remove-bag", removeBagHandler)
	r.Get("/api/{railType}/bags", listBagsHandler)
	r.Post("/api/{railType}/work-order", createWorkOrderHandler)
	r.Post("/api/{railType}/record-fix", recordFixHandler)
	r.Get("/api/{railType}/work-orders", listWorkOrdersHandler)

	r.Get("/health", healthHandler)

	// Serve global static files
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	port := ":8089"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	fmt.Println("Press Ctrl+C to stop")

	log.Fatal(http.ListenAndServe(port, r))
}
