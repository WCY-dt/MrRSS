package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/utils"
)

// ChatMessage represents a message in the chat conversation
type ChatMessage struct {
	Role    string `json:"role"` // "system", "user", or "assistant"
	Content string `json:"content"`
}

// ChatRequest represents the incoming chat request
type ChatRequest struct {
	Messages       []ChatMessage `json:"messages"`
	ArticleTitle   string        `json:"article_title,omitempty"`
	ArticleURL     string        `json:"article_url,omitempty"`
	ArticleContent string        `json:"article_content,omitempty"`
	IsFirstMessage bool          `json:"is_first_message,omitempty"`
}

// ChatResponse represents the response from the AI chat
type ChatResponse struct {
	Response string `json:"response"`
	HTML     string `json:"html,omitempty"` // Rendered HTML version of markdown response
}

// OpenAIRequest represents the request body for OpenAI-compatible APIs
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// OpenAIResponse represents the response from OpenAI-compatible APIs
type OpenAIResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// HandleAIChat handles chat requests for article discussions
// @Summary      AI chat with article
// @Description  Send messages to AI for discussing article content (requires ai_chat_enabled setting)
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        request  body      chat.ChatRequest  true  "Chat request (messages, article info)"
// @Success      200  {object}  chat.ChatResponse  "AI response (response, html)"
// @Failure      400  {object}  map[string]string  "Bad request (missing messages)"
// @Failure      403  {object}  map[string]string  "AI chat is disabled or limit reached"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat [post]
func HandleAIChat(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "Missing messages", http.StatusBadRequest)
		return
	}

	// Check if AI chat is enabled
	chatEnabled, _ := h.DB.GetSetting("ai_chat_enabled")
	if chatEnabled != "true" {
		http.Error(w, "AI chat is disabled", http.StatusForbidden)
		return
	}

	// Check if AI usage limit is reached
	if h.AITracker.IsLimitReached() {
		log.Printf("AI usage limit reached for chat")
		json.NewEncoder(w).Encode(map[string]string{
			"error": "AI usage limit reached",
		})
		return
	}

	// Apply rate limiting for AI requests
	h.AITracker.WaitForRateLimit()

	// Get AI settings
	apiKey, _ := h.DB.GetEncryptedSetting("ai_api_key")
	endpoint, _ := h.DB.GetSetting("ai_endpoint")
	model, _ := h.DB.GetSetting("ai_model")

	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	// Optimize context to reduce token usage
	optimizedMessages := optimizeChatContext(req.Messages, req.ArticleTitle, req.ArticleURL, req.ArticleContent, req.IsFirstMessage)

	// Try OpenAI format first
	response, err := tryOpenAIFormat(endpoint, apiKey, model, optimizedMessages, h)
	if err == nil {
		// Convert markdown response to HTML
		htmlResponse := utils.ConvertMarkdownToHTML(response)

		// Track AI usage (estimate tokens from input and output)
		estimatedTokens := estimateChatTokens(optimizedMessages, response)
		if err := h.AITracker.AddUsage(estimatedTokens); err != nil {
			log.Printf("Warning: failed to track AI usage: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Response: response, HTML: htmlResponse})
		return
	}

	// If OpenAI format fails, try Ollama format
	log.Printf("OpenAI format failed, trying Ollama format: %v", err)
	response, err = tryOllamaFormat(endpoint, apiKey, model, optimizedMessages, h)
	if err == nil {
		// Convert markdown response to HTML
		htmlResponse := utils.ConvertMarkdownToHTML(response)

		// Track AI usage (estimate tokens from input and output)
		estimatedTokens := estimateChatTokens(optimizedMessages, response)
		if err := h.AITracker.AddUsage(estimatedTokens); err != nil {
			log.Printf("Warning: failed to track AI usage: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Response: response, HTML: htmlResponse})
		return
	}

	// Both formats failed
	log.Printf("All chat formats failed: OpenAI error: %v, Ollama error: %v", err, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": "No response from AI"})
}

// createHTTPClientWithProxy creates an HTTP client with global proxy settings if enabled
func createHTTPClientWithProxy(h *core.Handler) (*http.Client, error) {
	// Check if global proxy is enabled
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled != "true" {
		return &http.Client{Timeout: 60 * time.Second}, nil
	}

	// Build proxy URL from global settings
	proxyType, _ := h.DB.GetSetting("proxy_type")
	proxyHost, _ := h.DB.GetSetting("proxy_host")
	proxyPort, _ := h.DB.GetSetting("proxy_port")
	proxyUsername, _ := h.DB.GetEncryptedSetting("proxy_username")
	proxyPassword, _ := h.DB.GetEncryptedSetting("proxy_password")

	// Build proxy URL
	proxyURL := buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)

	// Create HTTP client with proxy
	return createHTTPClient(proxyURL, 60*time.Second)
}

// buildProxyURL builds a proxy URL from components
func buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword string) string {
	if proxyHost == "" || proxyPort == "" {
		return ""
	}

	var urlBuilder strings.Builder
	urlBuilder.WriteString(strings.ToLower(proxyType))
	urlBuilder.WriteString("://")

	if proxyUsername != "" && proxyPassword != "" {
		urlBuilder.WriteString(url.QueryEscape(proxyUsername))
		urlBuilder.WriteString(":")
		urlBuilder.WriteString(url.QueryEscape(proxyPassword))
		urlBuilder.WriteString("@")
	}

	urlBuilder.WriteString(proxyHost)
	urlBuilder.WriteString(":")
	urlBuilder.WriteString(proxyPort)

	return urlBuilder.String()
}

// createHTTPClient creates an HTTP client with optional proxy
func createHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	client := &http.Client{Timeout: timeout}

	if proxyURL != "" {
		proxyFunc := http.ProxyFromEnvironment
		if proxyURL != "" {
			u, err := url.Parse(proxyURL)
			if err != nil {
				return nil, fmt.Errorf("invalid proxy URL: %w", err)
			}
			proxyFunc = http.ProxyURL(u)
		}
		client.Transport = &http.Transport{
			Proxy: proxyFunc,
		}
	}

	return client, nil
}

// isLocalEndpoint checks if a host is a local endpoint
func isLocalEndpoint(host string) bool {
	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		if !strings.Contains(host[idx:], "]") {
			host = host[:idx]
		}
	}
	// Remove brackets from IPv6 addresses
	host = strings.Trim(host, "[]")

	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		strings.HasPrefix(host, "127.") ||
		host == "0.0.0.0"
}

// tryOpenAIFormat attempts to use OpenAI-compatible API format for chat
func tryOpenAIFormat(endpoint, apiKey, model string, messages []ChatMessage, h *core.Handler) (string, error) {
	openAIReq := OpenAIRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1024,
	}

	jsonBody, err := json.Marshal(openAIReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	resp, err := sendChatRequest(endpoint, apiKey, jsonBody, h)
	if err != nil {
		return "", fmt.Errorf("OpenAI request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("OpenAI API returned status: %d", resp.StatusCode)
		if len(bodyBytes) > 0 {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, string(bodyBytes))
		}
		return "", fmt.Errorf("%s", errorMsg)
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	// Check for API error
	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 || openAIResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no response found in OpenAI response")
	}

	return strings.TrimSpace(openAIResp.Choices[0].Message.Content), nil
}

// tryOllamaFormat attempts to use Ollama API format for chat
func tryOllamaFormat(endpoint, apiKey, model string, messages []ChatMessage, h *core.Handler) (string, error) {
	// Convert messages to Ollama prompt format
	var promptBuilder strings.Builder
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			promptBuilder.WriteString("System: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		case "user":
			promptBuilder.WriteString("User: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		case "assistant":
			promptBuilder.WriteString("Assistant: ")
			promptBuilder.WriteString(msg.Content)
			promptBuilder.WriteString("\n\n")
		}
	}
	promptBuilder.WriteString("Assistant: ")

	requestBody := map[string]interface{}{
		"model":  model,
		"prompt": promptBuilder.String(),
		"stream": false,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	resp, err := sendChatRequest(endpoint, apiKey, jsonBody, h)
	if err != nil {
		return "", fmt.Errorf("Ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("Ollama API returned status: %d", resp.StatusCode)
		if len(bodyBytes) > 0 {
			errorMsg = fmt.Sprintf("%s - %s", errorMsg, string(bodyBytes))
		}
		return "", fmt.Errorf("%s", errorMsg)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	if !ollamaResp.Done || ollamaResp.Response == "" {
		return "", fmt.Errorf("no response found in Ollama response")
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// sendChatRequest sends the HTTP request for chat with proper headers and validation
func sendChatRequest(endpoint, apiKey string, jsonBody []byte, h *core.Handler) (*http.Response, error) {
	// Validate endpoint URL
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid API endpoint URL: %w", err)
	}

	// Both HTTP and HTTPS are allowed
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("API endpoint must use HTTP or HTTPS")
	}

	// Create HTTP client with proxy support if configured
	client, err := createHTTPClientWithProxy(h)
	if err != nil {
		log.Printf("Failed to create HTTP client with proxy: %v", err)
		client = &http.Client{Timeout: 60 * time.Second}
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	return client.Do(req)
}

// estimateChatTokens estimates token usage for chat requests
func estimateChatTokens(messages []ChatMessage, response string) int64 {
	var total int64
	for _, msg := range messages {
		// Rough estimation: 1 token ≈ 4 characters
		total += int64(len(msg.Content) / 4)
	}
	total += int64(len(response) / 4)
	return total
}

// optimizeChatContext optimizes the chat context to reduce token usage and manage context length
func optimizeChatContext(messages []ChatMessage, articleTitle, articleURL, articleContent string, isFirstMessage bool) []ChatMessage {
	const maxContextTokens = 8000 // Reserve tokens for response
	const maxArticleTokens = 2000 // Max tokens for article content
	const minArticleTokens = 500  // Min tokens to keep for context

	optimized := make([]ChatMessage, 0, len(messages))

	// Handle system message with article context
	if len(messages) > 0 && messages[0].Role == "system" {
		systemContent := messages[0].Content

		if isFirstMessage && articleContent != "" {
			// First message: include article context but limit length
			articleTokens := estimateTokens(articleContent)
			if articleTokens > maxArticleTokens {
				// Truncate article content intelligently
				articleContent = truncateArticleContent(articleContent, maxArticleTokens)
			}

			systemContent = fmt.Sprintf("You are a helpful AI assistant discussing an article with the user.\n\nArticle Title: %s\nArticle URL: %s\nArticle Content: %s\n\nPlease answer questions about this article. Be concise and helpful.\n\nIMPORTANT:\n- Respond in the SAME LANGUAGE as the user's message.\n- Use markdown formatting for better readability.", articleTitle, articleURL, articleContent)
		} else {
			// Subsequent messages: use minimal context
			systemContent = fmt.Sprintf("You are a helpful AI assistant discussing the article \"%s\".\n\nContinue the conversation about this article. Be concise and helpful.\n\nIMPORTANT:\n- Respond in the SAME LANGUAGE as the user's message.\n- Use markdown formatting for better readability.", articleTitle)
		}

		optimized = append(optimized, ChatMessage{
			Role:    "system",
			Content: systemContent,
		})

		messages = messages[1:] // Remove original system message
	} else if isFirstMessage && articleContent != "" {
		// No system message provided, create one for first message
		articleTokens := estimateTokens(articleContent)
		if articleTokens > maxArticleTokens {
			articleContent = truncateArticleContent(articleContent, maxArticleTokens)
		}

		systemContent := fmt.Sprintf("You are a helpful AI assistant discussing an article with the user.\n\nArticle Title: %s\nArticle URL: %s\nArticle Content: %s\n\nPlease answer questions about this article. Be concise and helpful.\n\nIMPORTANT:\n- Respond in the SAME LANGUAGE as the user's message.\n- Use markdown formatting for better readability.", articleTitle, articleURL, articleContent)

		optimized = append(optimized, ChatMessage{
			Role:    "system",
			Content: systemContent,
		})
	}

	// Process conversation messages with token-aware truncation
	conversationMessages := messages
	totalTokens := estimateTokens(getSystemContent(optimized))

	// Add messages from most recent backwards until we hit token limit
	for i := len(conversationMessages) - 1; i >= 0; i-- {
		msg := conversationMessages[i]
		msgTokens := estimateTokens(msg.Content)

		if totalTokens+msgTokens > maxContextTokens {
			// If we can't fit this message, try to summarize older messages
			if i > 0 { // Keep at least one message
				remainingTokens := maxContextTokens - totalTokens - 100 // Reserve some tokens
				if remainingTokens > minArticleTokens {
					// Add a summary of truncated messages
					summaryMsg := ChatMessage{
						Role:    "assistant",
						Content: fmt.Sprintf("[Previous conversation truncated to save tokens. %d messages omitted]", i+1),
					}
					optimized = append([]ChatMessage{summaryMsg}, optimized...)
				}
			}
			break
		}

		// Add message at the beginning (to maintain chronological order)
		optimized = append([]ChatMessage{msg}, optimized...)
		totalTokens += msgTokens
	}

	return optimized
}

// estimateTokens provides a rough token count estimation
func estimateTokens(text string) int {
	// Rough estimation: 1 token ≈ 4 characters for English text
	// This is a simplification; actual tokenization is more complex
	return len(text) / 4
}

// truncateArticleContent intelligently truncates article content to fit within token limit
func truncateArticleContent(content string, maxTokens int) string {
	if estimateTokens(content) <= maxTokens {
		return content
	}

	// Try to truncate at sentence boundaries
	sentences := strings.Split(content, ".")
	truncated := ""
	currentTokens := 0

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}
		sentence += "."

		sentenceTokens := estimateTokens(sentence)
		if currentTokens+sentenceTokens > maxTokens-100 { // Reserve tokens for truncation notice
			break
		}

		truncated += sentence + " "
		currentTokens += sentenceTokens
	}

	if len(truncated) < len(content) {
		truncated += "\n\n[Content truncated to save tokens]"
	}

	return strings.TrimSpace(truncated)
}

// getSystemContent extracts system message content for token counting
func getSystemContent(messages []ChatMessage) string {
	for _, msg := range messages {
		if msg.Role == "system" {
			return msg.Content
		}
	}
	return ""
}
