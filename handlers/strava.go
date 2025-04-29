package handlers

import (
	"encoding/json"
	"fmt"
	"joe-portal/backend/strava"
	"log"
	"net/http"
	"strings"
)

type StravaActivity struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance"`
	MovingTime  int     `json:"moving_time"`
	Type        string  `json:"type"`
	StartDate   string  `json:"start_date"`
}

func GetStravaActivities(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for Strava activities")

	// Check if we have an access token
	if strava.GetToken() == "" {
		log.Println("No access token available, initiating authorization flow")
		// No token available, return response to trigger authorization
		response := map[string]interface{}{
			"status":   "unauthorized",
			"message":  "Strava authorization required",
			"auth_url": strava.GetAuthorizationURL(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Using access token: %s", strava.GetToken())

	// We have a token, try to fetch activities
	activities, err := strava.FetchLatestPosts(strava.GetToken())
	if err != nil {
		log.Printf("Error fetching activities: %v", err)
		http.Error(w, "Failed to fetch Strava activities", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully fetched %d activities", len(activities))

	// Convert to our response format
	var response []StravaActivity
	for _, activity := range activities {
		response = append(response, StravaActivity{
			ID:          activity.ID,
			Name:        activity.Title,
			Description: activity.Description,
			Distance:    activity.Distance,
			MovingTime:  activity.MovingTime,
			Type:        activity.Type,
			StartDate:   activity.StartDate,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StravaAuth initiates the Strava OAuth flow
func StravaAuth(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Strava OAuth flow...")
	authURL := strava.GetAuthorizationURL()
	log.Printf("Generated authorization URL: %s", authURL)

	// Return the authorization URL to the client
	response := map[string]string{
		"auth_url": authURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StravaCallback handles the OAuth callback from Strava
func StravaCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Strava OAuth callback")

	// Get the authorization code and scope from the query parameters
	code := r.URL.Query().Get("code")
	scope := r.URL.Query().Get("scope")

	log.Printf("Authorization code received: %s", code)
	log.Printf("Granted scopes: %s", scope)

	if code == "" {
		log.Println("Error: Authorization code not found in callback")
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Check if we have the required scopes
	requiredScopes := map[string]bool{
		"activity:read":  true,
		"activity:write": true,
	}

	// Parse the scope string into individual scopes
	grantedScopes := make(map[string]bool)
	for _, s := range strings.Split(scope, ",") {
		grantedScopes[s] = true
	}

	log.Printf("Parsed granted scopes: %v", grantedScopes)

	// Check if all required scopes were granted
	missingScopes := []string{}
	for requiredScope := range requiredScopes {
		if !grantedScopes[requiredScope] {
			missingScopes = append(missingScopes, requiredScope)
		}
	}

	if len(missingScopes) > 0 {
		log.Printf("Error: Missing required scopes: %v", missingScopes)
		response := map[string]interface{}{
			"status":         "error",
			"message":        "Required scopes were not granted",
			"missing_scopes": missingScopes,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Println("All required scopes granted, proceeding with token exchange")

	// Exchange the code for tokens
	tokenResponse, err := strava.ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("Error exchanging code for tokens: %v", err)
		http.Error(w, fmt.Sprintf("Failed to exchange code for tokens: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully exchanged code for tokens. Access token: %s, Refresh token: %s",
		tokenResponse.AccessToken, tokenResponse.RefreshToken)
	log.Printf("Token expires in: %d seconds", tokenResponse.ExpiresIn)

	// Store the tokens in the TokenStore
	strava.SetToken(tokenResponse.AccessToken, tokenResponse.RefreshToken)

	// Return success response with token information
	response := map[string]interface{}{
		"status":        "success",
		"message":       "Successfully authorized with Strava",
		"access_token":  tokenResponse.AccessToken,
		"refresh_token": tokenResponse.RefreshToken,
		"expires_at":    tokenResponse.ExpiresAt,
		"expires_in":    tokenResponse.ExpiresIn,
		"scope":         scope,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateStravaActivity handles updating a Strava activity's title and description
func UpdateStravaActivity(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to update Strava activity")

	// Check if we have an access token
	if strava.GetToken() == "" {
		log.Println("No access token available")
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	// Parse the request body
	var update struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the activity using the Strava API
	err := strava.UpdateActivity(strava.GetToken(), update.ID, update.Title, update.Description)
	if err != nil {
		log.Printf("Error updating activity: %v", err)
		http.Error(w, "Failed to update activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Activity updated successfully",
	})
}
