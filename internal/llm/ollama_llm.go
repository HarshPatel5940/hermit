package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

// OllamaLLM handles text generation using Ollama.
type OllamaLLM struct {
	client *api.Client
	model  string
	logger *zap.Logger
}

// NewOllamaLLM creates a new Ollama LLM service.
func NewOllamaLLM(ollamaURL string, model string, logger *zap.Logger) *OllamaLLM {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		logger.Warn("Failed to create Ollama client from environment, using default", zap.Error(err))
		client = &api.Client{}
	}

	return &OllamaLLM{
		client: client,
		model:  model,
		logger: logger,
	}
}

// GenerateResponse generates a response from the LLM given a prompt.
func (l *OllamaLLM) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	req := &api.GenerateRequest{
		Model:  l.model,
		Prompt: prompt,
		Stream: new(bool), // Disable streaming for simple response
	}

	var fullResponse strings.Builder

	err := l.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fullResponse.WriteString(resp.Response)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	response := fullResponse.String()

	l.logger.Debug("Generated LLM response",
		zap.String("model", l.model),
		zap.Int("promptLength", len(prompt)),
		zap.Int("responseLength", len(response)),
	)

	return response, nil
}

// GenerateWithContext generates a response with context chunks from RAG.
func (l *OllamaLLM) GenerateWithContext(ctx context.Context, query string, contextChunks []string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query cannot be empty")
	}

	// Build prompt with context
	prompt := l.buildRAGPrompt(query, contextChunks)

	return l.GenerateResponse(ctx, prompt)
}

// buildRAGPrompt constructs a prompt for RAG-based generation.
func (l *OllamaLLM) buildRAGPrompt(query string, contextChunks []string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("You are a helpful assistant that answers questions based on the provided context.\n\n")

	if len(contextChunks) > 0 {
		promptBuilder.WriteString("Context:\n")
		for i, chunk := range contextChunks {
			promptBuilder.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, chunk))
		}
	}

	promptBuilder.WriteString(fmt.Sprintf("Question: %s\n\n", query))
	promptBuilder.WriteString("Answer the question based on the context provided above. ")
	promptBuilder.WriteString("If the context doesn't contain relevant information, say so. ")
	promptBuilder.WriteString("Be concise and accurate.\n\n")
	promptBuilder.WriteString("Answer: ")

	return promptBuilder.String()
}

// Chat performs a conversational chat with optional system message.
func (l *OllamaLLM) Chat(ctx context.Context, messages []ChatMessage, systemMessage string) (string, error) {
	// Convert messages to Ollama format
	var apiMessages []api.Message

	if systemMessage != "" {
		apiMessages = append(apiMessages, api.Message{
			Role:    "system",
			Content: systemMessage,
		})
	}

	for _, msg := range messages {
		apiMessages = append(apiMessages, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	req := &api.ChatRequest{
		Model:    l.model,
		Messages: apiMessages,
		Stream:   new(bool), // Disable streaming
	}

	var fullResponse strings.Builder

	err := l.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		fullResponse.WriteString(resp.Message.Content)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("chat failed: %w", err)
	}

	return fullResponse.String(), nil
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// GetModelInfo retrieves information about the current LLM model.
func (l *OllamaLLM) GetModelInfo(ctx context.Context) (*api.ShowResponse, error) {
	req := &api.ShowRequest{
		Model: l.model,
	}

	resp, err := l.client.Show(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}

	return resp, nil
}
