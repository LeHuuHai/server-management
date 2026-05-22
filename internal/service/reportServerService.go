package service

import (
	"context"
	"encoding/json"
	"fmt"
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
	aggregator *es.Aggregator
	exporter   export.ReportServerExporter
	publisher  mq.Publisher
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
		Path:     filePath,
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
	return s.publisher.Publish(ctx, "mail", mailReqByte)
}

func NewReportServerService(
	a *es.Aggregator,
	e export.ReportServerExporter,
	p mq.Publisher,
) *ReportServerService {
	return &ReportServerService{
		aggregator: a,
		exporter:   e,
		publisher:  p,
	}
}
