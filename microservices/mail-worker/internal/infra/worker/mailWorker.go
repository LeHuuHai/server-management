package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/domain/mail"
	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/domain/service"
)

type MailWorker struct {
	consumer mq.MailConsumerInterface
	sender   mail.SenderInterface
	download service.DownloadServiceInterface
}

func NewMailWorker(consumer mq.MailConsumerInterface, sender mail.SenderInterface, download service.DownloadServiceInterface) *MailWorker {
	return &MailWorker{
		consumer: consumer,
		sender:   sender,
		download: download,
	}
}

func (w *MailWorker) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, commitFunc, err := w.consumer.Read(ctx)
			if err != nil {
				slog.Warn("Failed to read mail request from Kafka", slog.Any("error", err))
				continue
			}

			slog.Info("Received mail request", "to", req.Mail.To)

			valid := true
			for i, attachment := range req.Mail.Attachments {
				data, err := w.download.Download(ctx, attachment.Filename)
				if err != nil {
					slog.Warn("Failed to download attachment", slog.String("filename", attachment.Filename), slog.Any("error", err))
					valid = false
					break
				}
				req.Mail.Attachments[i].Data = data
			}

			if valid {
				// Send the email
				err = w.sender.Send(ctx, req.Mail)
				if err != nil {
					slog.Warn("Failed to send email", slog.Any("to", req.Mail.To), slog.Any("error", err))
				} else {
					slog.Info("Successfully processed and sent email", "to", req.Mail.To)
				}
			} else {
				slog.Warn("Skipping email due to missing attachment(s)")
			}

			// Commit to Kafka regardless of outcome
			if commitFunc != nil {
				err = commitFunc(ctx)
				if err != nil {
					slog.Warn("Failed to commit mail request to Kafka", slog.Any("error", err))
				}
			}
		}
	}
}
