package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	agentconfig "github.com/LeHuuHai/server-management/config/agent"
	"github.com/LeHuuHai/server-management/internal/model"
)

func main() {
	ctx := context.Background()

	cfg, err := agentconfig.Load()
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	ticker := time.NewTicker(time.Duration(cfg.AppConfig.CycleHeartbeat) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sendHeartbeat(
				ctx,
				client,
				cfg.AppConfig.HeartbeatURL,
				cfg.AppConfig.ServerID,
				cfg.AppConfig.HeartbeatAPIKey,
			); err != nil {
				log.Printf("heartbeat failed: %v", err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func sendHeartbeat(
	ctx context.Context,
	client *http.Client,
	url string,
	serverID string,
	apiKey string,
) error {
	body, err := json.Marshal(model.Heartbeat{
		ServerID: serverID,
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
