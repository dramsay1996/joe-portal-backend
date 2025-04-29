package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// TokenStore holds the current Strava tokens
var TokenStore = struct {
	AccessToken  string
	RefreshToken string
}{
	AccessToken:  "",
	RefreshToken: "",
}

// GetToken returns the current access token
func GetToken() string {
	return TokenStore.AccessToken
}

// SetToken stores the tokens
func SetToken(accessToken, refreshToken string) {
	TokenStore.AccessToken = accessToken
	TokenStore.RefreshToken = refreshToken
}

// RemoveToken clears the tokens
func RemoveToken() {
	TokenStore.AccessToken = ""
	TokenStore.RefreshToken = ""
}

type Activity struct {
	ID          int64   `json:"id"`
	Title       string  `json:"name"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance"`
	MovingTime  int     `json:"moving_time"`
	Type        string  `json:"type"`
	StartDate   string  `json:"start_date"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func RefreshAccessToken(refreshToken string) (string, error) {
	url := "https://www.strava.com/oauth/token"
	payload := map[string]string{
		"client_id":     os.Getenv("STRAVA_CLIENT_ID"),
		"client_secret": os.Getenv("STRAVA_CLIENT_SECRET"),
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"scope":         "activity:read,activity:write",
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var refreshResp refreshResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&refreshResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Received new access token: %s and refresh token: %s\n", refreshResp.AccessToken, refreshResp.RefreshToken)
	fmt.Printf("Token type: %s, Expires in: %d seconds\n", refreshResp.TokenType, refreshResp.ExpiresIn)

	// Update in-memory token store
	TokenStore.AccessToken = refreshResp.AccessToken
	TokenStore.RefreshToken = refreshResp.RefreshToken

	return refreshResp.AccessToken, nil
}

func FetchLatestPosts(accessToken string) ([]Activity, error) {
	fmt.Println("FETCHING ACTIVITIES")
	url := "https://www.strava.com/api/v3/athlete/activities?per_page=10"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	fmt.Printf("Making request to Strava API with token: Bearer %s\n", accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest posts: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Strava API response status: %d\n", resp.StatusCode)
	if resp.StatusCode == http.StatusUnauthorized {
		// Token might be expired; attempt refresh
		fmt.Println("Strava responded with 401, attempting token refresh...")

		// Use the existing RefreshAccessToken function
		newAccessToken, err := RefreshAccessToken(TokenStore.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}

		// Retry the original request with new token
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request after refresh: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+newAccessToken)

		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch posts after refresh: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, fmt.Errorf("unexpected status code after refresh: %d, body: %s", resp.StatusCode, string(body))
		}
	} else if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var activities []Activity
	err = json.Unmarshal(body, &activities)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return activities, nil
}

// GetAuthorizationURL returns the URL to redirect users to for Strava authorization
func GetAuthorizationURL() string {
	clientID := os.Getenv("STRAVA_CLIENT_ID")
	// Use the frontend URL for redirect
	redirectURI := os.Getenv("FRONTEND_URL") + "/api/strava/callback"

	// Construct the authorization URL with required scopes
	authURL := fmt.Sprintf(
		"https://www.strava.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=activity:read,activity:write",
		clientID,
		redirectURI,
	)

	return authURL
}

// ExchangeCodeForToken exchanges an authorization code for access and refresh tokens
func ExchangeCodeForToken(code string) (*refreshResponse, error) {
	url := "https://www.strava.com/oauth/token"
	payload := map[string]string{
		"client_id":     os.Getenv("STRAVA_CLIENT_ID"),
		"client_secret": os.Getenv("STRAVA_CLIENT_SECRET"),
		"code":          code,
		"grant_type":    "authorization_code",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResponse refreshResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &tokenResponse, nil
}

// UpdateActivity updates a Strava activity's title and description
func UpdateActivity(accessToken string, activityID int64, newTitle string, newDescription string) error {
	url := fmt.Sprintf("https://www.strava.com/api/v3/activities/%d", activityID)

	payload := map[string]string{
		"name":        newTitle,
		"description": newDescription,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Token might be expired; attempt refresh
		fmt.Println("Strava responded with 401, attempting token refresh...")

		// Use the existing RefreshAccessToken function
		newAccessToken, err := RefreshAccessToken(TokenStore.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to refresh token: %v", err)
		}

		// Retry the original request with new token
		req, err = http.NewRequest("PUT", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create request after refresh: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+newAccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to update activity after refresh: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("unexpected status code after refresh: %d, body: %s", resp.StatusCode, string(body))
		}
	} else if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
