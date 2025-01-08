package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type Config struct {
    BotToken string
    ChatID   string
}

type TelegramMessage struct {
    ChatID string `json:"chat_id"`
    Text   string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type MessageRequest struct {
    Message string `json:"message"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

func sendTelegramMessage(config Config, message string) error {
    baseURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken)
    
    telegramMsg := TelegramMessage{
        ChatID: config.ChatID,
        Text:   message,
		ParseMode: "HTML",
    }
    
    jsonData, err := json.Marshal(telegramMsg)
    if err != nil {
        return fmt.Errorf("error marshaling message: %v", err)
    }
    
    resp, err := http.Post(
        baseURL,
        "application/json",
        strings.NewReader(string(jsonData)),
    )
    if err != nil {
        return fmt.Errorf("error sending message: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    return nil
}

func handleSendMessage(w http.ResponseWriter, r *http.Request, config Config) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req MessageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
        return
    }
    
    if req.Message == "" {
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Message cannot be empty"})
        return
    }
    
    err := sendTelegramMessage(config, req.Message)
    if err != nil {
        json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{"status": "Message sent successfully"})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}
	
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    chatID := os.Getenv("TELEGRAM_CHAT_ID")
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{allowedOrigins},
		AllowCredentials: true,
	})

	mux := http.NewServeMux()

	handler := c.Handler(mux)
    
    if botToken == "" || chatID == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID environment variables are required")
    }
    
    config := Config{
        BotToken: botToken,
        ChatID:   chatID,
    }
    
    mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
        handleSendMessage(w, r, config)
    })
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "4000"
    }
    
    fmt.Printf("Server running on port %s...\n", port)
    if err := http.ListenAndServe(":"+port, c.Handler(handler)); err != nil {
        log.Fatal(err)
    }
}
