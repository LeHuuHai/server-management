package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/mail"
	"os"

	"github.com/LeHuuHai/server-management/internal/domain/file/export"
	"github.com/LeHuuHai/server-management/internal/domain/mq"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	es "github.com/LeHuuHai/server-management/internal/infra/elasticsearch"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/google/uuid"
)

type ReportServerService struct {
	aggregator *es.CachedAggregator
	exporter   export.ReportServerExporter
	publisher  mq.Publisher
	mailTopic  string
}

func (s *ReportServerService) ReportServer(ctx context.Context, request model.GenServerReportRequest) error {
	// valid
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
	// aggregation
	report, err := s.aggregator.Aggregation(ctx, request.From, request.To)
	if err != nil {
		return err
	}
	// export file
	fileName := fmt.Sprintf("report-%s.%s", uuid.NewString(), s.exporter.FileType())
	filePath := fmt.Sprintf("./tmp/%s", fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	err = s.exporter.Export(ctx, file, report)
	if err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	// publish req mail
	attachment := make([]model.Attachment, 0)
	attachment = append(attachment, model.Attachment{
		Filename: fileName,
		Data:     []byte{},
	})
	mailReq := model.RequestMail{
		Mail: model.Mail{
			From:        "", // depend sender
			To:          request.Receivers,
			Subject:     "Server uptime report",
			Body:        "Please find the attached report.",
			Attachments: attachment,
		},
	}
	mailReqByte, err := json.Marshal(mailReq)
	if err != nil {
		return err
	}
	msg := mq.Message{
		Topic: s.mailTopic,
		Value: mailReqByte,
	}
	err = s.publisher.Publish(ctx, msg)
	if err != nil {
		return err
	}
	slog.Info("Report generated and mail request published", "file_path", filePath, "receivers", request.Receivers)
	return nil
}

func NewReportServerService(
	a *es.CachedAggregator,
	e export.ReportServerExporter,
	p mq.Publisher,
	mailTopic string,
) *ReportServerService {
	os.MkdirAll("./tmp", 0755)
	return &ReportServerService{
		aggregator: a,
		exporter:   e,
		publisher:  p,
		mailTopic:  mailTopic,
	}
}
