package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	bedrockruntime "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type LLMRequest struct {
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
	Temperature  float64
}

type StructuredAnalysis struct {
	Summary      string   `json:"summary"`
	LikelyIssue  string   `json:"likely_issue"`
	Confidence   float64  `json:"confidence"`
	Evidence     []string `json:"evidence"`
	PotentialFix []string `json:"potential_fix"`
	NextChecks   []string `json:"next_checks"`
}

type ProviderResult struct {
	Provider   string              `json:"provider"`
	Type       string              `json:"type"`
	Model      string              `json:"model"`
	DurationMS int64               `json:"duration_ms"`
	Response   string              `json:"response,omitempty"`
	Parsed     *StructuredAnalysis `json:"parsed,omitempty"`
	Error      string              `json:"error,omitempty"`
}

type LLMProvider interface {
	Name() string
	Type() string
	Model() string
	PrepareRequest(req LLMRequest) LLMRequest
	Complete(ctx context.Context, req LLMRequest) (string, error)
}

func buildProviders(backends []BackendConfig) ([]LLMProvider, error) {
	providers := make([]LLMProvider, 0, len(backends))
	for _, backend := range backends {
		provider, err := buildProvider(backend)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func buildProvider(cfg BackendConfig) (LLMProvider, error) {
	switch cfg.Type {
	case "", "openai":
		return newOpenAIProvider(cfg)
	case "ollama":
		return newOllamaProvider(cfg)
	case "bedrock":
		return newBedrockProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported backend type %q", cfg.Type)
	}
}

type openAIProvider struct {
	name         string
	model        string
	baseURL      string
	apiKey       string
	systemPrompt string
	maxTokens    int
	temperature  float64
	httpClient   *http.Client
}

func newOpenAIProvider(cfg BackendConfig) (LLMProvider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("openai backend %q is missing model", cfg.Name)
	}

	apiKey := ""
	if cfg.APIKeyEnv != "" {
		apiKey = strings.TrimSpace(os.Getenv(cfg.APIKeyEnv))
	}
	if apiKey == "" {
		return nil, fmt.Errorf("openai backend %q is missing API key env %q", cfg.Name, cfg.APIKeyEnv)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &openAIProvider{
		name:         cfg.Name,
		model:        cfg.Model,
		baseURL:      strings.TrimRight(baseURL, "/"),
		apiKey:       apiKey,
		systemPrompt: cfg.SystemPrompt,
		maxTokens:    cfg.MaxTokens,
		temperature:  cfg.Temperature,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *openAIProvider) Name() string  { return p.name }
func (p *openAIProvider) Type() string  { return "openai" }
func (p *openAIProvider) Model() string { return p.model }
func (p *openAIProvider) PrepareRequest(req LLMRequest) LLMRequest {
	return applyProviderOverrides(req, p.systemPrompt, p.maxTokens, p.temperature)
}

func (p *openAIProvider) Complete(ctx context.Context, req LLMRequest) (string, error) {
	payload := map[string]any{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.UserPrompt},
		},
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build openai request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read openai response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

type ollamaProvider struct {
	name         string
	model        string
	baseURL      string
	systemPrompt string
	maxTokens    int
	temperature  float64
	httpClient   *http.Client
}

func newOllamaProvider(cfg BackendConfig) (LLMProvider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("ollama backend %q is missing model", cfg.Name)
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://ollama:11434"
	}
	return &ollamaProvider{
		name:         cfg.Name,
		model:        cfg.Model,
		baseURL:      strings.TrimRight(baseURL, "/"),
		systemPrompt: cfg.SystemPrompt,
		maxTokens:    cfg.MaxTokens,
		temperature:  cfg.Temperature,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *ollamaProvider) Name() string  { return p.name }
func (p *ollamaProvider) Type() string  { return "ollama" }
func (p *ollamaProvider) Model() string { return p.model }
func (p *ollamaProvider) PrepareRequest(req LLMRequest) LLMRequest {
	return applyProviderOverrides(req, p.systemPrompt, p.maxTokens, p.temperature)
}

func (p *ollamaProvider) Complete(ctx context.Context, req LLMRequest) (string, error) {
	payload := map[string]any{
		"model":  p.model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.UserPrompt},
		},
		"options": map[string]any{
			"temperature": req.Temperature,
			"num_predict": req.MaxTokens,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read ollama response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}
	return strings.TrimSpace(parsed.Message.Content), nil
}

type bedrockProvider struct {
	name         string
	model        string
	region       string
	systemPrompt string
	maxTokens    int
	temperature  float64
}

func newBedrockProvider(cfg BackendConfig) (LLMProvider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("bedrock backend %q is missing model", cfg.Name)
	}
	region := cfg.Region
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_REGION"))
	}
	if region == "" {
		return nil, fmt.Errorf("bedrock backend %q is missing region", cfg.Name)
	}
	return &bedrockProvider{
		name:         cfg.Name,
		model:        cfg.Model,
		region:       region,
		systemPrompt: cfg.SystemPrompt,
		maxTokens:    cfg.MaxTokens,
		temperature:  cfg.Temperature,
	}, nil
}

func (p *bedrockProvider) Name() string  { return p.name }
func (p *bedrockProvider) Type() string  { return "bedrock" }
func (p *bedrockProvider) Model() string { return p.model }
func (p *bedrockProvider) PrepareRequest(req LLMRequest) LLMRequest {
	return applyProviderOverrides(req, p.systemPrompt, p.maxTokens, p.temperature)
}

func (p *bedrockProvider) Complete(ctx context.Context, req LLMRequest) (string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(p.region))
	if err != nil {
		return "", fmt.Errorf("load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	payload := map[string]any{
		"anthropic_version": "bedrock-2023-05-31",
		"messages": []map[string]string{
			{"role": "user", "content": req.UserPrompt},
		},
		"max_tokens": req.MaxTokens,
	}
	if req.SystemPrompt != "" {
		payload["system"] = req.SystemPrompt
	}
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal bedrock request: %w", err)
	}

	output, err := client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(p.model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		return "", fmt.Errorf("bedrock invoke failed: %w", err)
	}

	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(output.Body, &parsed); err != nil {
		return strings.TrimSpace(string(output.Body)), nil
	}

	var parts []string
	for _, block := range parsed.Content {
		if block.Text != "" {
			parts = append(parts, block.Text)
		}
	}
	if len(parts) == 0 {
		return strings.TrimSpace(string(output.Body)), nil
	}
	return strings.TrimSpace(strings.Join(parts, "\n")), nil
}

func applyProviderOverrides(req LLMRequest, systemPrompt string, maxTokens int, temperature float64) LLMRequest {
	if strings.TrimSpace(systemPrompt) != "" {
		req.SystemPrompt = systemPrompt
	}
	if maxTokens > 0 {
		req.MaxTokens = maxTokens
	}
	if temperature > 0 {
		req.Temperature = temperature
	}
	return req
}
