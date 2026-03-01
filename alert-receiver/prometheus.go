package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type PrometheusClient struct {
	baseURL    string
	httpClient *http.Client
}

type MetricSnapshot struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Query       string         `json:"query"`
	ResultType  string         `json:"result_type,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	Series      []MetricSeries `json:"series,omitempty"`
	Error       string         `json:"error,omitempty"`
}

type MetricSeries struct {
	Labels map[string]string `json:"labels,omitempty"`
	Value  string            `json:"value"`
}

func NewPrometheusClient(baseURL string, timeout time.Duration) *PrometheusClient {
	return &PrometheusClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *PrometheusClient) InstantQuery(ctx context.Context, query MetricQuery, queryTime time.Time) (MetricSnapshot, error) {
	params := url.Values{}
	params.Set("query", query.Query)
	params.Set("time", queryTime.Format(time.RFC3339))

	endpoint := p.baseURL + "/api/v1/query?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return MetricSnapshot{}, fmt.Errorf("build Prometheus request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return MetricSnapshot{}, fmt.Errorf("query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MetricSnapshot{}, fmt.Errorf("read Prometheus response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return MetricSnapshot{}, fmt.Errorf("Prometheus status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string          `json:"resultType"`
			Result     json.RawMessage `json:"result"`
		} `json:"data"`
		ErrorType string `json:"errorType"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return MetricSnapshot{}, fmt.Errorf("decode Prometheus response: %w", err)
	}
	if apiResp.Status != "success" {
		return MetricSnapshot{}, fmt.Errorf("Prometheus %s: %s", apiResp.ErrorType, apiResp.Error)
	}

	snapshot := MetricSnapshot{
		Name:        query.Name,
		Description: query.Description,
		Query:       query.Query,
		ResultType:  apiResp.Data.ResultType,
	}

	switch apiResp.Data.ResultType {
	case "scalar":
		var raw []any
		if err := json.Unmarshal(apiResp.Data.Result, &raw); err != nil {
			return MetricSnapshot{}, fmt.Errorf("decode scalar result: %w", err)
		}
		if len(raw) == 2 {
			snapshot.Series = []MetricSeries{{Value: fmt.Sprint(raw[1])}}
			snapshot.Summary = fmt.Sprintf("value=%s", fmt.Sprint(raw[1]))
		}
	case "vector":
		var entries []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"`
		}
		if err := json.Unmarshal(apiResp.Data.Result, &entries); err != nil {
			return MetricSnapshot{}, fmt.Errorf("decode vector result: %w", err)
		}
		for _, entry := range entries {
			value := ""
			if len(entry.Value) == 2 {
				value = fmt.Sprint(entry.Value[1])
			}
			snapshot.Series = append(snapshot.Series, MetricSeries{
				Labels: entry.Metric,
				Value:  value,
			})
		}
		snapshot.Summary = summarizeSeries(snapshot.Series)
	default:
		snapshot.Summary = string(apiResp.Data.Result)
	}

	return snapshot, nil
}

func summarizeSeries(series []MetricSeries) string {
	if len(series) == 0 {
		return "no series"
	}

	parts := make([]string, 0, len(series))
	for _, s := range series {
		labelParts := make([]string, 0, len(s.Labels))
		for k, v := range s.Labels {
			if k == "__name__" {
				continue
			}
			labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(labelParts)
		if len(labelParts) == 0 {
			parts = append(parts, s.Value)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s => %s", strings.Join(labelParts, ","), s.Value))
	}
	return strings.Join(parts, "; ")
}
