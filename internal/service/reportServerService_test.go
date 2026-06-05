package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

type mockReportAggregator struct {
	result []model.ServerUptimeAgg
	err    error

	called bool
	from   time.Time
	to     time.Time
}

func (m *mockReportAggregator) Aggregation(ctx context.Context, from time.Time, to time.Time) ([]model.ServerUptimeAgg, error) {
	m.called = true
	m.from = from
	m.to = to
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockReportExporter struct {
	fileType    string
	contentType string
	err         error

	called bool
	data   []model.ServerUptimeAgg
}

func (m *mockReportExporter) Export(ctx context.Context, writer io.Writer, data []model.ServerUptimeAgg) error {
	m.called = true
	m.data = data
	if m.err != nil {
		return m.err
	}
	_, err := writer.Write([]byte("report data"))
	return err
}

func (m *mockReportExporter) FileType() string {
	if m.fileType == "" {
		return "xlsx"
	}
	return m.fileType
}

func (m *mockReportExporter) ContentType() string {
	if m.contentType == "" {
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	return m.contentType
}

type mockReportPublisher struct {
	err error

	called bool
	msg    mq.Message
}

func (m *mockReportPublisher) Publish(ctx context.Context, msg mq.Message) error {
	m.called = true
	m.msg = msg
	return m.err
}

func withTempWorkDir(t *testing.T) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}

func TestReportServer_Success(t *testing.T) {
	withTempWorkDir(t)

	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	report := []model.ServerUptimeAgg{
		{ServerID: "s1", UptimeRatio: 99.5, DocCount: 10},
	}
	agg := &mockReportAggregator{result: report}
	exporter := &mockReportExporter{}
	publisher := &mockReportPublisher{}
	svc := service.NewReportServerService(agg, exporter, publisher, "mail-topic")

	err := svc.ReportServer(context.Background(), model.GenServerReportRequest{
		From:      from,
		To:        to,
		Receivers: []string{"ops@example.com", "admin@example.com"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !agg.called {
		t.Fatal("expected aggregator to be called")
	}
	if !agg.from.Equal(from) || !agg.to.Equal(to) {
		t.Errorf("expected aggregation range %v-%v, got %v-%v", from, to, agg.from, agg.to)
	}
	if !exporter.called {
		t.Fatal("expected exporter to be called")
	}
	if len(exporter.data) != 1 || exporter.data[0].ServerID != "s1" {
		t.Errorf("expected exporter to receive report data, got %+v", exporter.data)
	}
	if !publisher.called {
		t.Fatal("expected publisher to be called")
	}
	if publisher.msg.Topic != "mail-topic" {
		t.Errorf("expected topic mail-topic, got %s", publisher.msg.Topic)
	}

	var mailReq model.RequestMail
	if err := json.Unmarshal(publisher.msg.Value, &mailReq); err != nil {
		t.Fatalf("unmarshal published mail request: %v", err)
	}
	if !equalStrings(mailReq.Mail.To, []string{"ops@example.com", "admin@example.com"}) {
		t.Errorf("expected receivers to be published, got %+v", mailReq.Mail.To)
	}
	if mailReq.Mail.Subject != "Server uptime report" {
		t.Errorf("unexpected mail subject: %s", mailReq.Mail.Subject)
	}
	if len(mailReq.Mail.Attachments) != 1 {
		t.Fatalf("expected one attachment, got %d", len(mailReq.Mail.Attachments))
	}
	attachment := mailReq.Mail.Attachments[0]
	if !strings.HasPrefix(attachment.Filename, "report-") || !strings.HasSuffix(attachment.Filename, ".xlsx") {
		t.Errorf("unexpected attachment filename: %s", attachment.Filename)
	}
	if len(attachment.Data) != 0 {
		t.Errorf("expected attachment data to be empty in mail request, got %d bytes", len(attachment.Data))
	}

	reportPath := filepath.Join("tmp", attachment.Filename)
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("expected report file to be created: %v", err)
	}
	if !bytes.Equal(content, []byte("report data")) {
		t.Errorf("unexpected report file content: %q", string(content))
	}
}

func TestReportServer_InvalidTimeRange(t *testing.T) {
	withTempWorkDir(t)

	agg := &mockReportAggregator{}
	exporter := &mockReportExporter{}
	publisher := &mockReportPublisher{}
	svc := service.NewReportServerService(agg, exporter, publisher, "mail-topic")

	err := svc.ReportServer(context.Background(), model.GenServerReportRequest{
		From:      time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Receivers: []string{"ops@example.com"},
	})

	if !errors.Is(err, apperr.ErrInvalidTimeRange) {
		t.Errorf("expected ErrInvalidTimeRange, got %v", err)
	}
	if agg.called || exporter.called || publisher.called {
		t.Error("dependencies should not be called for invalid time range")
	}
}

func TestReportServer_InvalidEmail(t *testing.T) {
	cases := []struct {
		name      string
		receivers []string
	}{
		{name: "empty receivers", receivers: nil},
		{name: "malformed email", receivers: []string{"not-an-email"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			withTempWorkDir(t)

			agg := &mockReportAggregator{}
			exporter := &mockReportExporter{}
			publisher := &mockReportPublisher{}
			svc := service.NewReportServerService(agg, exporter, publisher, "mail-topic")

			err := svc.ReportServer(context.Background(), model.GenServerReportRequest{
				From:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
				To:        time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
				Receivers: tt.receivers,
			})

			if !errors.Is(err, apperr.ErrInvalidEmail) {
				t.Errorf("expected ErrInvalidEmail, got %v", err)
			}
			if agg.called || exporter.called || publisher.called {
				t.Error("dependencies should not be called for invalid email")
			}
		})
	}
}

func TestReportServer_AggregatorError(t *testing.T) {
	withTempWorkDir(t)

	aggErr := errors.New("aggregation failed")
	agg := &mockReportAggregator{err: aggErr}
	exporter := &mockReportExporter{}
	publisher := &mockReportPublisher{}
	svc := service.NewReportServerService(agg, exporter, publisher, "mail-topic")

	err := svc.ReportServer(context.Background(), validReportRequest())

	if !errors.Is(err, aggErr) {
		t.Errorf("expected aggregator error, got %v", err)
	}
	if exporter.called || publisher.called {
		t.Error("exporter and publisher should not be called when aggregation fails")
	}
}

func TestReportServer_ExporterError(t *testing.T) {
	withTempWorkDir(t)

	exportErr := errors.New("export failed")
	agg := &mockReportAggregator{result: []model.ServerUptimeAgg{{ServerID: "s1"}}}
	exporter := &mockReportExporter{err: exportErr}
	publisher := &mockReportPublisher{}
	svc := service.NewReportServerService(agg, exporter, publisher, "mail-topic")

	err := svc.ReportServer(context.Background(), validReportRequest())

	if !errors.Is(err, exportErr) {
		t.Errorf("expected exporter error, got %v", err)
	}
	if publisher.called {
		t.Error("publisher should not be called when export fails")
	}
}

func TestReportServer_PublisherError(t *testing.T) {
	withTempWorkDir(t)

	publishErr := errors.New("publish failed")
	agg := &mockReportAggregator{result: []model.ServerUptimeAgg{{ServerID: "s1"}}}
	exporter := &mockReportExporter{}
	publisher := &mockReportPublisher{err: publishErr}
	svc := service.NewReportServerService(agg, exporter, publisher, "mail-topic")

	err := svc.ReportServer(context.Background(), validReportRequest())

	if !errors.Is(err, publishErr) {
		t.Errorf("expected publisher error, got %v", err)
	}
	if !exporter.called {
		t.Error("exporter should be called before publisher failure")
	}
}

func validReportRequest() model.GenServerReportRequest {
	return model.GenServerReportRequest{
		From:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
		Receivers: []string{"ops@example.com"},
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
