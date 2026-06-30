package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LeHuuHai/server-management/microservices/agent/config"
	"github.com/LeHuuHai/server-management/microservices/agent/internal/model"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	ticker := time.NewTicker(time.Duration(cfg.AppConfig.CycleHeartbeat) * time.Millisecond)
	defer ticker.Stop()

	slog.Info("Agent started, sending heartbeats...", "server_id", cfg.AppConfig.ServerID, "interval_ms", cfg.AppConfig.CycleHeartbeat)

	go func() {
		// Send initial heartbeat immediately on startup
		if err := sendHeartbeat(ctx, client, cfg.AppConfig.HeartbeatURL, cfg.AppConfig.ServerID, cfg.AppConfig.HeartbeatAPIKey); err != nil {
			slog.Warn("Initial heartbeat failed", slog.Any("err", err))
		} else {
			slog.Info("Initial heartbeat sent successfully")
		}

		for {
			select {
			case <-ticker.C:
				if err := sendHeartbeat(ctx, client, cfg.AppConfig.HeartbeatURL, cfg.AppConfig.ServerID, cfg.AppConfig.HeartbeatAPIKey); err != nil {
					slog.Warn("Send heartbeat failed", slog.Any("err", err))
				} else {
					slog.Info("Sent heartbeat")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Block until a shutdown signal is received
	<-sigChan
	slog.Info("Shutdown signal received, stopping agent gracefully...")
	cancel()

	// Give the goroutine a small moment to exit cleanly
	time.Sleep(100 * time.Millisecond)
	slog.Info("Agent stopped successfully")
}

func sendHeartbeat(
	ctx context.Context,
	client *http.Client,
	url string,
	serverID string,
	apiKey string,
) error {
	body, err := json.Marshal(model.Heartbeat{
		ServerID:  serverID,
		Timestamp: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
