package service_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LeHuuHai/server-management/internal/service"
)

func TestDownloadService_Success(t *testing.T) {
	content := []byte("report data")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer ts.Close()

	svc := service.NewDownLoadService(ts.URL, "secret", ts.Client())
	data, err := svc.Download(context.Background(), "report.xlsx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", content, data)
	}
}

func TestDownloadService_Non200Status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	svc := service.NewDownLoadService(ts.URL, "key", ts.Client())
	_, err := svc.Download(context.Background(), "missing.xlsx")
	if err == nil {
		t.Error("expected error on 404 status")
	}
}

func TestDownloadService_APIKeyIsSet(t *testing.T) {
	var receivedKey string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	svc := service.NewDownLoadService(ts.URL, "my-api-key", ts.Client())
	_, _ = svc.Download(context.Background(), "file.xlsx")

	if receivedKey != "my-api-key" {
		t.Errorf("expected API key my-api-key, got %s", receivedKey)
	}
}

func TestDownloadService_InvalidURL(t *testing.T) {
	svc := service.NewDownLoadService("http://localhost:0", "key", &http.Client{})
	_, err := svc.Download(context.Background(), "file.xlsx")
	if err == nil {
		t.Error("expected connection error")
	}
}

func TestDownloadService_ContextCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	svc := service.NewDownLoadService(ts.URL, "key", ts.Client())
	_, err := svc.Download(ctx, "file.xlsx")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
