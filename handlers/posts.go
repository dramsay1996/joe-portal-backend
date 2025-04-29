package handlers

import (
	"encoding/json"
	"net/http"
)

type StravaPost struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// GetLatestStravaPosts retrieves the most recent Strava activities
func GetLatestStravaPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement Strava API integration to fetch latest activities
	// This is a mock response for now
	posts := []StravaPost{
		{ID: "1", Title: "Morning Run"},
		{ID: "2", Title: "Evening Ride"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// UpdateStravaPostTitle updates the title of a specific Strava activity
func UpdateStravaPostTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var post StravaPost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: Implement Strava API integration to update activity title
	// This is a mock response for now
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}
