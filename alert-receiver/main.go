package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type analysisJob struct {
	ID         string
	ReceivedAt time.Time
	Payload    GrafanaWebhookPayload
}

type analysisRecord struct {
	ID             string            `json:"id"`
	ReceivedAt     time.Time         `json:"received_at"`
	CompletedAt    time.Time         `json:"completed_at"`
	AlertStatus    string            `json:"alert_status"`
	Receiver       string            `json:"receiver"`
	GroupKey       string            `json:"group_key"`
	CommonLabels   map[string]string `json:"common_labels"`
	CommonAnnots   map[string]string `json:"common_annotations"`
	AlertSummaries []alertSummary    `json:"alerts"`
	Metrics        []MetricSnapshot  `json:"metrics,omitempty"`
	Providers      []ProviderResult  `json:"providers,omitempty"`
	Error          string            `json:"error,omitempty"`
}

type alertSummary struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      time.Time         `json:"ends_at"`
}

type analysisStore struct {
	max   int
	items []analysisRecord
	mu    sync.RWMutex
}

func newAnalysisStore(max int) *analysisStore {
	return &analysisStore{max: max}
}

func (s *analysisStore) add(record analysisRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append([]analysisRecord{record}, s.items...)
	if len(s.items) > s.max {
		s.items = s.items[:s.max]
	}
}

func (s *analysisStore) list() []analysisRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]analysisRecord, len(s.items))
	copy(out, s.items)
	return out
}

type server struct {
	cfg       Config
	prom      *PrometheusClient
	providers []LLMProvider
	queue     chan analysisJob
	store     *analysisStore
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	registerMetrics()

	providers, err := buildProviders(cfg.Backends)
	if err != nil {
		slog.Error("failed to build providers", "error", err)
		os.Exit(1)
	}

	promClient := NewPrometheusClient(cfg.PrometheusURL, cfg.PrometheusTimeout)
	srv := &server{
		cfg:       cfg,
		prom:      promClient,
		providers: providers,
		queue:     make(chan analysisJob, cfg.JobQueueSize),
		store:     newAnalysisStore(cfg.MaxStoredAnalyses),
	}

	for i := 0; i < cfg.WorkerCount; i++ {
		go srv.worker(i + 1)
	}

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           srv.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("starting alert-receiver",
		"port", cfg.Port,
		"prometheus_url", cfg.PrometheusURL,
		"backends", providerNames(providers),
		"workers", cfg.WorkerCount,
	)

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleHealthz)
	mux.HandleFunc("/alerts/grafana", s.handleGrafanaWebhook)
	mux.HandleFunc("/analyses/latest", s.handleLatestAnalyses)
	return mux
}

func (s *server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          "ok",
		"providers":       providerNames(s.providers),
		"prometheus_url":  s.cfg.PrometheusURL,
		"queue_depth":     len(s.queue),
		"worker_count":    s.cfg.WorkerCount,
		"stored_analyses": len(s.store.list()),
	})
}

func (s *server) handleLatestAnalyses(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.store.list(),
	})
}

func (s *server) handleGrafanaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var payload GrafanaWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	alertsReceivedTotal.WithLabelValues(payload.Status).Inc()

	job := analysisJob{
		ID:         fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeID(payload.GroupKey)),
		ReceivedAt: time.Now().UTC(),
		Payload:    payload,
	}

	select {
	case s.queue <- job:
		queueDepthGauge.Inc()
		slog.Info("alert queued",
			"job_id", job.ID,
			"receiver", payload.Receiver,
			"status", payload.Status,
			"alerts", len(payload.Alerts),
		)
		writeJSON(w, http.StatusAccepted, map[string]any{
			"job_id":   job.ID,
			"status":   "queued",
			"alerts":   len(payload.Alerts),
			"backends": providerNames(s.providers),
		})
	default:
		jobResultsTotal.WithLabelValues("queue_full").Inc()
		http.Error(w, "queue full", http.StatusServiceUnavailable)
	}
}

func (s *server) worker(id int) {
	for job := range s.queue {
		queueDepthGauge.Dec()
		s.processJob(id, job)
	}
}

func (s *server) processJob(workerID int, job analysisJob) {
	start := time.Now()
	record := analysisRecord{
		ID:             job.ID,
		ReceivedAt:     job.ReceivedAt,
		AlertStatus:    job.Payload.Status,
		Receiver:       job.Payload.Receiver,
		GroupKey:       job.Payload.GroupKey,
		CommonLabels:   job.Payload.CommonLabels,
		CommonAnnots:   job.Payload.CommonAnnotations,
		AlertSummaries: summarizeAlerts(job.Payload.Alerts),
	}

	slog.Info("processing alert job",
		"job_id", job.ID,
		"worker", workerID,
		"alerts", len(job.Payload.Alerts),
	)

	metrics, err := s.collectMetrics(job)
	if err != nil {
		record.Error = err.Error()
		slog.Warn("metric collection failed", "job_id", job.ID, "error", err)
	}
	record.Metrics = metrics

	if len(s.providers) == 0 {
		record.Providers = []ProviderResult{{
			Provider: "none",
			Type:     "none",
			Error:    "no LLM backends configured",
		}}
	} else {
		record.Providers = s.runProviders(job, metrics)
	}

	record.CompletedAt = time.Now().UTC()
	jobDurationSeconds.Observe(time.Since(start).Seconds())
	jobResultsTotal.WithLabelValues("processed").Inc()
	s.store.add(record)

	slog.Info("alert job completed",
		"job_id", job.ID,
		"worker", workerID,
		"duration", time.Since(start).String(),
	)
}

func (s *server) collectMetrics(job analysisJob) ([]MetricSnapshot, error) {
	if strings.TrimSpace(s.cfg.PrometheusURL) == "" {
		return nil, nil
	}

	queryTime := time.Now().UTC()
	if len(job.Payload.Alerts) > 0 {
		earliest := earliestAlertTime(job.Payload, queryTime)
		queryTime = earliest.Add(s.cfg.PrometheusLookback)
		if queryTime.After(time.Now().UTC()) {
			queryTime = time.Now().UTC()
		}
	}

	snapshots := make([]MetricSnapshot, 0, len(s.cfg.MetricQueries))
	for _, query := range s.cfg.MetricQueries {
		snapshot, err := s.prom.InstantQuery(context.Background(), query, queryTime)
		if err != nil {
			prometheusQueriesTotal.WithLabelValues(query.Name, "error").Inc()
			snapshots = append(snapshots, MetricSnapshot{
				Name:        query.Name,
				Description: query.Description,
				Query:       query.Query,
				Error:       err.Error(),
			})
			continue
		}
		prometheusQueriesTotal.WithLabelValues(query.Name, "success").Inc()
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func (s *server) runProviders(job analysisJob, metrics []MetricSnapshot) []ProviderResult {
	request, err := buildLLMRequest(job, metrics, s.cfg.PrometheusLookback)
	if err != nil {
		return []ProviderResult{{
			Provider: "prompt-builder",
			Type:     "internal",
			Error:    err.Error(),
		}}
	}

	results := make([]ProviderResult, len(s.providers))
	var wg sync.WaitGroup
	for i, provider := range s.providers {
		wg.Add(1)
		go func(idx int, provider LLMProvider) {
			defer wg.Done()
			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), s.cfg.LLMTimeout)
			defer cancel()

			response, err := provider.Complete(ctx, provider.PrepareRequest(request))
			durationMS := time.Since(start).Milliseconds()

			result := ProviderResult{
				Provider:   provider.Name(),
				Type:       provider.Type(),
				Model:      provider.Model(),
				DurationMS: durationMS,
			}

			if err != nil {
				providerRequestsTotal.WithLabelValues(provider.Name(), "error").Inc()
				result.Error = err.Error()
				results[idx] = result
				return
			}

			providerRequestsTotal.WithLabelValues(provider.Name(), "success").Inc()
			result.Response = response

			var parsed StructuredAnalysis
			if err := json.Unmarshal([]byte(response), &parsed); err == nil && parsed.Summary != "" {
				result.Parsed = &parsed
			}

			results[idx] = result
		}(i, provider)
	}
	wg.Wait()
	return results
}

func summarizeAlerts(alerts []GrafanaAlert) []alertSummary {
	out := make([]alertSummary, 0, len(alerts))
	for _, alert := range alerts {
		out = append(out, alertSummary{
			Status:      alert.Status,
			Labels:      alert.Labels,
			Annotations: alert.Annotations,
			StartsAt:    alert.StartsAt,
			EndsAt:      alert.EndsAt,
		})
	}
	return out
}

func providerNames(providers []LLMProvider) []string {
	names := make([]string, 0, len(providers))
	for _, provider := range providers {
		names = append(names, provider.Name())
	}
	sort.Strings(names)
	return names
}

func sanitizeID(v string) string {
	replacer := strings.NewReplacer("/", "-", ":", "-", " ", "-", "\n", "-", "\t", "-")
	out := replacer.Replace(strings.TrimSpace(v))
	if out == "" {
		return "alert"
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
