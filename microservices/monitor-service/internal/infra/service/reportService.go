package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"os"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/aggregator"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/infra/file"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/google/uuid"
)

type ReportService struct {
	aggregator aggregator.ReportAggregator
	exporter   *file.ReportExporter
	publisher  mq.MailPublisherInterface
}

func NewReportService(
	agg aggregator.ReportAggregator,
	publisher mq.MailPublisherInterface,
) service.ReportServiceInterface {
	_ = os.MkdirAll("./tmp", 0755)
	return &ReportService{
		aggregator: agg,
		exporter:   file.NewReportExporter(),
		publisher:  publisher,
	}
}

func (s *ReportService) ReportServer(ctx context.Context, request model.GenServerReportRequest) error {
	if request.From.After(request.To) {
		return apperr.ErrInvalidTimeRange
	}
	if len(request.Receivers) == 0 {
		return apperr.ErrInvalidEmail
	}
	for _, email := range request.Receivers {
		if _, err := mail.ParseAddress(email); err != nil {
			return apperr.ErrInvalidEmail
		}
	}

	var report []model.ServerUptimeAgg
	var err error

	report, err = s.aggregator.Aggregation(ctx, request.From, request.To)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("report-%s.%s", uuid.NewString(), s.exporter.FileType())
	filePath := fmt.Sprintf("./tmp/%s", fileName)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = s.exporter.Export(ctx, f, report)
	if err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}

	// Publish RequestMail event to Kafka
	attachments := []pkgmodel.Attachment{
		{
			Filename: fileName,
			Data:     []byte{}, // Data is empty because the consumer pulls it via gRPC
		},
	}
	mailReq := pkgmodel.RequestMail{
		Mail: pkgmodel.Mail{
			From:        "",
			To:          request.Receivers,
			Subject:     "Server uptime report",
			Body:        "Please find the attached report.",
			Attachments: attachments,
		},
	}

	err = s.publisher.Publish(ctx, mailReq)
	if err != nil {
		return err
	}

	slog.Info("Report generated and mail request published to Kafka", "file_path", filePath, "receivers", request.Receivers)
	return nil
}
