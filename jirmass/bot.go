package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// BotConfig holds configuration for the automated customer bot.
type BotConfig struct {
	AdminOnline    bool
	AwayMessage    string
	OperatingHours string
	DjangoAPIURL   string
	LLMEnabled     bool
	LLMModel       string
	LLMAPIKey      string
	WebhookURL     string
}

var (
	botConfigLock sync.RWMutex
	currentConfig = BotConfig{
		AdminOnline:    false, // Admin is absent by default, activating auto-reply bot
		AwayMessage:    "The sales admin is currently offline. Our automated bot is active 24/7 to assist with water packaging rolls & bags orders.",
		OperatingHours: "Monday to Friday, 8:00 AM – 5:00 PM (GMT+1). Automated support is 24/7.",
		DjangoAPIURL:   "http://localhost:8000/api/products",
		LLMEnabled:     os.Getenv("GEMINI_API_KEY") != "" || os.Getenv("LLM_API_KEY") != "",
		LLMModel:       getEnvDefault("LLM_MODEL", "gemini-1.5-flash"),
		LLMAPIKey:      getEnvDefault("GEMINI_API_KEY", os.Getenv("LLM_API_KEY")),
		WebhookURL:     os.Getenv("ADMIN_WEBHOOK_URL"),
	}

	inquiriesLock   sync.Mutex
	loggedInquiries []CustomerInquiryRecord
)

func getEnvDefault(key, defVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defVal
}

// CustomerInquiryRecord tracks messages left by customers via the bot
type CustomerInquiryRecord struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Message     string    `json:"message"`
	BotResponse string    `json:"bot_response"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChatRequest represents an incoming customer message to the bot
type ChatRequest struct {
	Message string `json:"message"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// ChatResponse represents the automated reply from the bot
type ChatResponse struct {
	Reply            string   `json:"reply"`
	AdminOnline      bool     `json:"admin_online"`
	IsAwayReply      bool     `json:"is_away_reply"`
	InquiryLogged    bool     `json:"inquiry_logged"`
	IsLLMGenerated   bool     `json:"is_llm_generated"`
	SuggestedPrompts []string `json:"suggested_prompts"`
	Timestamp        string   `json:"timestamp"`
}

// BotStatusResponse reports current admin presence & bot health
type BotStatusResponse struct {
	AdminOnline    bool   `json:"admin_online"`
	BotActive      bool   `json:"bot_active"`
	AwayMessage    string `json:"away_message"`
	OperatingHours string `json:"operating_hours"`
	LLMEnabled     bool   `json:"llm_enabled"`
	LLMModel       string `json:"llm_model"`
	ServerTime     string `json:"server_time"`
}

// ProcessBotMessage analyzes customer input and generates automated responses focused on water packaging
func ProcessBotMessage(req ChatRequest) ChatResponse {
	botConfigLock.RLock()
	adminOnline := currentConfig.AdminOnline
	awayMessage := currentConfig.AwayMessage
	apiKey := currentConfig.LLMAPIKey
	model := currentConfig.LLMModel
	webhookURL := currentConfig.WebhookURL
	botConfigLock.RUnlock()

	userMsg := strings.TrimSpace(req.Message)
	lower := strings.ToLower(userMsg)

	now := time.Now().Format("15:04")
	resp := ChatResponse{
		AdminOnline: adminOnline,
		IsAwayReply: !adminOnline,
		Timestamp:   now,
		SuggestedPrompts: []string{
			"Water Packing Rolls info",
			"Water Packing Bags info",
			"Place a direct order",
			"Admin status",
		},
	}

	// Case 1: Check if this is a lead / direct contact registration
	if req.Email != "" || req.Phone != "" || strings.Contains(lower, "@") {
		email := req.Email
		if email == "" && strings.Contains(lower, "@") {
			for _, part := range strings.Fields(userMsg) {
				if strings.Contains(part, "@") {
					email = strings.Trim(part, ".,;:()")
					break
				}
			}
		}

		name := req.Name
		if name == "" {
			name = "Pure Water Factory Partner"
		}

		replyText := fmt.Sprintf(
			"Thank you %s! Because our sales administration team is currently offline, I have logged your order inquiry and contact details (%s). Our production representative will contact you via WhatsApp/Phone shortly with factory pricing and delivery timelines.",
			name, email,
		)

		logInquiryLocally(name, email, req.Phone, userMsg, replyText)
		forwardInquiryToDjango(name, email, req.Phone, userMsg, replyText)
		dispatchLiveWebhookAlert("customer_inquiry", fmt.Sprintf("Water Packaging Lead: %s", name), map[string]string{
			"Name":    name,
			"Email":   email,
			"Phone":   req.Phone,
			"Message": userMsg,
		}, webhookURL)

		resp.Reply = replyText
		resp.InquiryLogged = true
		return resp
	}

	// Case 2: Inquiries about admin, human agent, or presence
	if containsAny(lower, "admin", "human", "person", "agent", "manager", "staff", "offline", "online", "who are you") {
		if !adminOnline {
			resp.Reply = fmt.Sprintf(
				"Notice: Our human sales managers are currently AWAY (%s). I am Jirmass Bot, your automated assistant. I can help you with water packing rolls & bags specifications, plain vs printed options, and order placement. To request a callback, simply leave your factory name, phone number, and requirements!",
				awayMessage,
			)
		} else {
			resp.Reply = "Our sales admin is currently online. How can we assist your pure water factory today? You can place an order directly or ask about roll specifications."
		}
		return resp
	}

	// Case 3: Try LLM Integration if API Key is configured
	if apiKey != "" {
		llmReply, err := queryLLM(userMsg, apiKey, model)
		if err == nil && len(strings.TrimSpace(llmReply)) > 0 {
			resp.Reply = llmReply
			resp.IsLLMGenerated = true
			resp.SuggestedPrompts = []string{"Place Direct Order", "Pricing per kg", "Delivery details"}
			return resp
		}
	}

	// Case 4: Water Packing Bags (outer bags / bundling sacks)
	if containsAny(lower, "bag", "bags", "packing bag", "outer bag", "bundle", "bale", "sack", "20 sachets") {
		resp.Reply = "🛍️ Jirmass Water Packing Bags (Outer Bundling Bags):\n\n" +
			"• Heavy-duty outer bags engineered to bundle and hold 20 sachets of pure water.\n" +
			"• High puncture resistance and tear strength for rough transport and stacking.\n" +
			"• Available in Plain and Custom Printed branding.\n" +
			"• Packaged in convenient bundles (e.g., 500 bags per bundle, 1,000 bags, or 5,000 bag bales).\n\n" +
			"👉 You can place a direct order for bags and rolls using the form on this page!"
		resp.SuggestedPrompts = []string{"Water Packing Rolls info", "Place Direct Order", "Delivery timeframe"}
		return resp
	}

	// Case 5: Water Packing Rolls (pure water / sachet packaging film)
	if containsAny(lower, "roll", "rolls", "film", "sachet", "pure water", "320mm", "print", "printed", "gauge", "micron", "machine") {
		resp.Reply = "💧 Jirmass Water Packing Rolls:\n\n" +
			"• Standard 320mm roll width, calibrated for automated form-fill-seal packaging machines.\n" +
			"• 100% virgin food-grade LDPE with high tensile strength for leak-free sealing.\n" +
			"• Custom Printed Rolls: High-definition rotogravure printing with your factory brand, NAFDAC number, and batch codes.\n" +
			"• Plain / Unprinted Rolls: High-clarity transparent film rolls.\n" +
			"• Available by roll count or bulk weight (kg).\n\n" +
			"👉 To order, use the Direct Order form on this page or reply with your factory name, location, and quantity needed!"
		resp.SuggestedPrompts = []string{"Water Packing Bags info", "Place Direct Order", "How to get a quote"}
		return resp
	}

	// Case 6: Direct Order Placement / Pricing inquiries
	if containsAny(lower, "order", "price", "cost", "quote", "rate", "how much", "buy", "wholesale", "direct") {
		resp.Reply = "Factory-Direct Wholesale Pricing:\n\n" +
			"Our water packaging rolls and bags are priced competitively per roll, per kg, or per bundle based on volume and print requirements.\n\n" +
			"To place your direct order immediately:\n" +
			"1. Scroll to the 'Direct Order' section on this page.\n" +
			"2. Select your product (Rolls, Bags, or Both) and specify quantity.\n" +
			"3. Provide your Phone/WhatsApp and Delivery address.\n\n" +
			"Our sales team receives the alert instantly and will contact you to confirm production and delivery!"
		resp.SuggestedPrompts = []string{"Leave my phone number", "Roll specifications", "Working hours"}
		return resp
	}

	// Case 7: Delivery / Location / Dispatch
	if containsAny(lower, "delivery", "shipping", "deliver", "ship", "location", "address", "where", "factory", "transport") {
		resp.Reply = "🚚 Factory Dispatch & Logistics:\n\n" +
			"We dispatch water packaging rolls and packing bags directly to water factories and regional depots. Standard delivery turnaround is 1–3 business days for plain rolls/bags and scheduled production runs for custom branded prints. Direct factory loading is also available."
		return resp
	}

	// Case 8: Working Hours
	if containsAny(lower, "hour", "hours", "time", "open", "schedule", "when") {
		botConfigLock.RLock()
		hours := currentConfig.OperatingHours
		botConfigLock.RUnlock()
		resp.Reply = fmt.Sprintf("Factory operations and sales hours: %s. Our automated bot is available 24/7 to take your orders while staff is away.", hours)
		return resp
	}

	// Case 9: Greetings
	if containsAny(lower, "hello", "hi", "hey", "good morning", "good afternoon", "good evening", "greetings") {
		resp.Reply = "👋 Welcome to Jirmass Plastics! We specialize exclusively in Water Packing Rolls and Water Packing Bags for pure water factories. Our admin is currently away, but you can ask about roll specs, outer bags, or place your direct order right here. How can I assist you?"
		return resp
	}

	// Case 10: Closing / Gratitude
	if containsAny(lower, "thank", "thanks", "bye", "goodbye") {
		resp.Reply = "You are most welcome! Thank you for choosing Jirmass Plastics for your water packaging supplies. Reach out anytime if you need more rolls or packing bags!"
		return resp
	}

	// Default Fallback
	resp.Reply = "Welcome to Jirmass Plastics — manufacturers of Water Packing Rolls and Water Packing Bags. Since our sales manager is currently offline, you can place a direct order using the form on this page, or reply with your WhatsApp/phone number and requirements to receive a fast callback."
	return resp
}

func queryLLM(prompt, apiKey, model string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	systemPrompt := "You are Jirmass Bot, the 24/7 automated sales assistant for Jirmass Plastics. " +
		"We specialize EXCLUSIVELY in two products for pure water factories: " +
		"1) Water Packing Rolls (pure water sachet packaging film rolls, standard 320mm, virgin LDPE, custom printed or plain, leak-free sealing) and " +
		"2) Water Packing Bags (outer bundling bags holding 20 sachets of water, high-strength LDPE, bundles/bales). " +
		"Do NOT discuss unrelated plastics like chairs, crates, buckets, or jerrycans. " +
		"The human sales administrator is currently offline/away. Answer questions concisely and professionally, and guide the user to place a direct order."

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": systemPrompt + "\nCustomer inquiry: " + prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API returned %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text), nil
	}

	return "", fmt.Errorf("no candidates in LLM response")
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func logInquiryLocally(name, email, phone, message, botResponse string) {
	inquiriesLock.Lock()
	defer inquiriesLock.Unlock()

	id := len(loggedInquiries) + 1
	loggedInquiries = append(loggedInquiries, CustomerInquiryRecord{
		ID:          id,
		Name:        name,
		Email:       email,
		Phone:       phone,
		Message:     message,
		BotResponse: botResponse,
		CreatedAt:   time.Now(),
	})
}

func forwardInquiryToDjango(name, email, phone, message, botResponse string) {
	go func() {
		payload := map[string]string{
			"name":         name,
			"email":        email,
			"phone":        phone,
			"message":      message,
			"bot_response": botResponse,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return
		}

		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Post("http://localhost:8000/api/products/inquiries/", "application/json", bytes.NewBuffer(data))
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
}

func dispatchLiveWebhookAlert(event, title string, details map[string]string, webhookURL string) {
	if webhookURL == "" {
		return
	}
	go func() {
		payload := map[string]interface{}{
			"event":        event,
			"title":        title,
			"details":      details,
			"triggered_at": time.Now().Format(time.RFC3339),
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return
		}
		client := &http.Client{Timeout: 4 * time.Second}
		resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Webhook alert dispatch failed: %v", err)
			return
		}
		_ = resp.Body.Close()
	}()
}
