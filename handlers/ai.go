package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type BruleQuoteResponse struct {
	Quote string `json:"quote"`
}

type ContentCheckResponse struct {
	IsVulgar bool   `json:"isVulgar"`
	Message  string `json:"message"`
}

type Handler struct {
	openAIKey string
}

func NewHandler(openAIKey string) *Handler {
	return &Handler{
		openAIKey: openAIKey,
	}
}

// GenerateBruleQuote generates a Steve Brule quote using OpenAI
func GenerateBruleQuote(w http.ResponseWriter, r *http.Request) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	resp, err := client.CreateChatCompletion(
		r.Context(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Generate a single Steve Brule quote without quotation marks. Max 140 characters",
				},
			},
		},
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating quote: %v", err), http.StatusInternalServerError)
		return
	}

	quote := strings.TrimSpace(resp.Choices[0].Message.Content)
	if len(quote) > 140 {
		quote = quote[:140]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BruleQuoteResponse{
		Quote: quote,
	})
}

// CheckContentForVulgarity checks if content contains vulgar language and sends email if needed
func CheckContentForVulgarity(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Silently return success if we can't decode the request
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ContentCheckResponse{
			IsVulgar: false,
			Message:  "Content checked successfully",
		})
		return
	}

	// Initialize OpenAI client
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		fmt.Println("Warning: OPENAI_API_KEY not set")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ContentCheckResponse{
			IsVulgar: false,
			Message:  "Content checked successfully",
		})
		return
	}

	client := openai.NewClient(openAIKey)

	// Create the chat completion request
	resp, err := client.CreateChatCompletion(
		r.Context(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a content moderator. Your task is to check if text contains vulgar, offensive, or inappropriate language. Respond with only 'yes' or 'no'.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: request.Content,
				},
			},
		},
	)

	// Default to not vulgar if there's any error
	isVulgar := false
	if err == nil && len(resp.Choices) > 0 {
		response := strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content))
		isVulgar = response == "yes"
		fmt.Printf("Content check result for '%s': %s\n", request.Content, response)
	} else if err != nil {
		fmt.Printf("Error checking content: %v\n", err)
	}

	// Send email notification if vulgar content is detected
	if isVulgar {
		adminEmail := os.Getenv("ADMIN_EMAIL")
		sendgridKey := os.Getenv("SENDGRID_API_KEY")
		fmt.Println("Sending email to admin")
		fmt.Println(adminEmail)
		fmt.Println(sendgridKey)

		if adminEmail != "" && sendgridKey != "" {
			fmt.Println("Attempting to send email notification...")
			// Use a verified single sender email instead of a domain
			from := mail.NewEmail("Dustin", "dustinramsay2025@gmail.com") // Replace with your verified sender email
			to := mail.NewEmail("Admin", adminEmail)
			subject := "Vulgar Content Detected"
			plainTextContent := fmt.Sprintf("Vulgar content was detected in an activity: %s", request.Content)
			htmlContent := fmt.Sprintf("<strong>Vulgar content was detected in an activity:</strong><br>%s", request.Content)
			message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
			client := sendgrid.NewSendClient(sendgridKey)
			response, err := client.Send(message)
			if err != nil {
				fmt.Printf("Error sending email: %v\n", err)
				fmt.Printf("SendGrid response: %+v\n", response)
				if response != nil && response.StatusCode == 403 {
					fmt.Println("SendGrid 403 Error - This usually means either:")
					fmt.Println("1. The API key doesn't have 'Mail Send' permissions")
					fmt.Println("2. The sender email address isn't verified with SendGrid")
					fmt.Println("Please check your SendGrid account settings and verify your sender email.")
				}
			} else {
				fmt.Println("Email sent successfully")
				fmt.Printf("SendGrid response status code: %d\n", response.StatusCode)
			}
		} else {
			fmt.Println("Email not sent - missing environment variables:")
			if adminEmail == "" {
				fmt.Println("- ADMIN_EMAIL is not set")
			}
			if sendgridKey == "" {
				fmt.Println("- SENDGRID_API_KEY is not set")
			}
		}
	}

	// Always return success to the user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ContentCheckResponse{
		IsVulgar: false,
		Message:  "Content checked successfully",
	})
}

func (h *Handler) CheckContent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	client := openai.NewClient(h.openAIKey)
	resp, err := client.CreateChatCompletion(
		r.Context(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a content moderator. Your task is to check if text contains vulgar, offensive, or inappropriate language. Respond with only 'yes' or 'no'.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: request.Content,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("OpenAI API error: %v\n", err)
		http.Error(w, "Failed to check content", http.StatusInternalServerError)
		return
	}

	if len(resp.Choices) == 0 {
		http.Error(w, "No response from AI", http.StatusInternalServerError)
		return
	}

	response := resp.Choices[0].Message.Content
	isVulgar := strings.TrimSpace(strings.ToLower(response)) == "yes"

	json.NewEncoder(w).Encode(map[string]bool{
		"isVulgar": isVulgar,
	})
}
