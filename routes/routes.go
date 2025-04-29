package routes

import (
	"joe-portal/backend/handlers"
	"joe-portal/backend/middleware"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply CORS middleware to all routes
	router.Use(middleware.CORSMiddleware)

	// Public routes (no auth required)
	public := router.PathPrefix("/api").Subrouter()
	public.HandleFunc("/login", handlers.Login).Methods("POST", "OPTIONS")

	// Strava OAuth routes
	public.HandleFunc("/strava/auth", handlers.StravaAuth).Methods("GET", "OPTIONS")
	public.HandleFunc("/strava/callback", handlers.StravaCallback).Methods("GET", "OPTIONS")

	// Protected routes (auth required)
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// Strava routes
	protected.HandleFunc("/strava/activities", handlers.GetStravaActivities).Methods("GET", "OPTIONS")
	protected.HandleFunc("/strava/activities", handlers.UpdateStravaActivity).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/auth/logout", handlers.Logout).Methods("POST", "OPTIONS")

	// AI routes
	protected.HandleFunc("/ai/brule-quote", handlers.GenerateBruleQuote).Methods("GET", "OPTIONS")
	protected.HandleFunc("/ai/check-content", handlers.CheckContentForVulgarity).Methods("POST", "OPTIONS")

	return router
}
