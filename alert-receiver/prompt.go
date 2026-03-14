package main

import (
	"encoding/json"
	"fmt"
	"time"
)

const defaultSystemPrompt = `You analyze edge network alerts using only the provided evidence.
Return strict JSON with this shape:
{
  "summary": "short incident summary",
  "likely_issue": "most likely root cause",
  "confidence": 0.0,
  "evidence": ["bullet evidence"],
  "potential_fix": ["ordered remediation ideas"],
  "next_checks": ["additional checks if evidence is insufficient"]
}
Do not invent radio-level evidence if it is not present in the metrics.`

func buildLLMRequest(job analysisJob, metrics []MetricSnapshot, lookbackDuration time.Duration) (LLMRequest, error) {
	payload := map[string]any{
		"received_at":        job.ReceivedAt,
		"alert_status":       job.Payload.Status,
		"receiver":           job.Payload.Receiver,
		"group_key":          job.Payload.GroupKey,
		"group_labels":       job.Payload.GroupLabels,
		"common_labels":      job.Payload.CommonLabels,
		"common_annotations": job.Payload.CommonAnnotations,
		"alerts":             summarizeAlerts(job.Payload.Alerts),
		"metric_snapshots":   metrics,
		"analysis_window":    fmt.Sprint(lookbackDuration),
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return LLMRequest{}, fmt.Errorf("marshal prompt payload: %w", err)
	}

	return LLMRequest{
		SystemPrompt: defaultSystemPrompt,
		UserPrompt:   "Evaluate this Grafana alert incident and summarize the issue, likely cause, and potential fix using only the evidence below.\n\n" + string(body),
		MaxTokens:    900,
		Temperature:  0.2,
	}, nil
}
