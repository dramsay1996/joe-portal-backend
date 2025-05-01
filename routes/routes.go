package routes

import (
	"joe-portal/backend/handlers"
	"joe-portal/backend/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply logging middleware first
	router.Use(middleware.LoggingMiddleware)

	// Apply CORS middleware to all routes
	router.Use(middleware.CORSMiddleware)

	// Add root route for debugging
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "API is running"}`))
	})

	// Create API router with /api prefix
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Public routes (no auth required)
	apiRouter.HandleFunc("/login", handlers.Login).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/strava/auth", handlers.StravaAuth).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/strava/callback", handlers.StravaCallback).Methods("GET", "OPTIONS")

	// Protected routes (auth required)
	protected := apiRouter.PathPrefix("").Subrouter() // Empty prefix since we're already under /api
	protected.Use(middleware.AuthMiddleware)

	// Strava routes
	protected.HandleFunc("/strava/activities", handlers.GetStravaActivities).Methods("GET", "OPTIONS")
	protected.HandleFunc("/strava/activities", handlers.UpdateStravaActivity).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/auth/logout", handlers.Logout).Methods("POST", "OPTIONS")

	// AI routes
	protected.HandleFunc("/ai/brule-quote", handlers.GenerateBruleQuote).Methods("GET", "OPTIONS")
	protected.HandleFunc("/ai/check-content", handlers.CheckContentForVulgarity).Methods("POST", "OPTIONS")

	// Debug: Print all registered routes
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		log.Printf("Route: %s, Methods: %v", t, methods)
		return nil
	})

	return router
}
