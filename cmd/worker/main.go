package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	workerconfig "github.com/LeHuuHai/server-management/config/worker"
	"github.com/LeHuuHai/server-management/internal/domain/mail"
	"github.com/LeHuuHai/server-management/internal/domain/mq"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	smtp "github.com/LeHuuHai/server-management/internal/infra/mail"
	workerruntime "github.com/LeHuuHai/server-management/internal/infra/runtime/worker"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
	probing "github.com/prometheus-community/pro-bing"
	"gopkg.in/gomail.v2"
)

func CheckServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *workerruntime.App,
	consumer mq.Consumer,
	publisher mq.Publisher,
) {
	defer wg.Done()
	jobs := make(chan model.RequestPing, 10)
	var workerWG sync.WaitGroup
	workerWG.Add(rt.Config.AppConfig.NumThread + 1)
	go func() {
		defer workerWG.Done()
		defer close(jobs)
		for {
			// read req
			msg, err := consumer.Read(ctx)
			if err != nil {
				slog.Warn("Read topic ping failed", slog.Any("err", err))
				continue
			}
			consumer.Commit(ctx, msg)
			var pingReq model.RequestPing
			err = json.Unmarshal(msg.Value, &pingReq)
			if err != nil {
				slog.Warn("Marshal request ping failed", slog.Any("bytes", msg.Value), slog.Any("err", err))
				continue
			}
			slog.Info("Received ping request", "server_id", pingReq.ServerID, "ip", pingReq.IP)
			select {
			case jobs <- pingReq:
			case <-ctx.Done():
				return
			}
		}
	}()
	for i := 0; i < rt.Config.AppConfig.NumThread; i++ {
		go func() {
			defer workerWG.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case req, ok := <-jobs:
					if !ok {
						return
					}
					res := model.ResponsePing{
						ServerID: req.ServerID,
						Status:   string(model.StatusOffline),
						PingAt:   time.Now(),
					}
					pinger, err := probing.NewPinger(req.IP)
					if err == nil {
						pinger.Count = 1                 // Chỉ gửi đúng 1 gói ICMP duy nhất
						pinger.Timeout = 1 * time.Second // Timeout đúng 1 giây theo thiết kế của bạn

						// QUAN TRỌNG: Chế độ Privileged (Raw Socket) bắt buộc phải bật
						// để Linux không bắt ép dùng cổng Unprivileged UDP lằng nhằng.
						pinger.SetPrivileged(true)

						err = pinger.Run()
						if err == nil {
							// Kiểm tra kết quả thống kê gói tin
							stats := pinger.Statistics()
							if stats.PacketsRecv > 0 {
								res.Status = string(model.StatusOnline)
							}
						}

					}

					resBytes, err := json.Marshal(res)
					if err != nil {
						slog.Warn("Marshal response ping failed", slog.Any("response ping", res), slog.Any("err", err))
						continue
					}
					msg := mq.Message{
						Topic: rt.Config.KafkaConfig.Topics["ping_res"],
						Value: resBytes,
					}
					err = publisher.Publish(ctx, msg)
					if err != nil {
						slog.Warn("Publish response ping failed", slog.Any("response ping", res), slog.Any("err", err))
					}
					slog.Info("Processed ping request", "server_id", req.ServerID, "status", res.Status)
				}
			}
		}()
	}
	workerWG.Wait()
}

func SendMail(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *workerruntime.App,
	consumer mq.Consumer,
	sender mail.Sender,
	downloadService *service.DownloadService,
) {
	defer wg.Done()
	for {
		msg, err := consumer.Read(ctx)
		if err != nil {
			slog.Warn("Read topic mail failed", slog.Any("err", err))
			continue
		}
		var mailReq model.RequestMail
		err = json.Unmarshal(msg.Value, &mailReq)
		if err != nil {
			slog.Warn("Marshal request mail failed", slog.Any("request mail", mailReq), slog.Any("err", err))
			continue
		}
		slog.Info("Received mail request", "to", mailReq.Mail.To, "filename", mailReq.Mail.Attachments[0].Filename)
		// attachments
		valid := true
		for i, attachment := range mailReq.Mail.Attachments {
			data, err := downloadService.Download(
				ctx,
				attachment.Filename,
			)
			if err != nil {
				slog.Warn("Download report file failed", slog.Any("file name", attachment.Filename), slog.Any("err", err))
				valid = false
				break
			}

			mailReq.Mail.Attachments[i].Data = data
		}
		if !valid {
			slog.Warn("Cannot send mail because of miss attachment", slog.Any("file name", mailReq.Mail.Attachments))
			continue
		}
		err = sender.Send(ctx, mailReq.Mail)
		if err != nil {
			slog.Warn("Send mail failed", slog.Any("to", mailReq.Mail.To), slog.Any("filename", mailReq.Mail.Attachments[0].Filename), slog.Any("err", err))
			continue
		}
		consumer.Commit(ctx, msg)
		slog.Info("Processed mail request", "to", mailReq.Mail.To, "filename", mailReq.Mail.Attachments[0].Filename)
	}
}

func main() {
	ctx := context.Background()

	cfg, err := workerconfig.Load()
	if err != nil {
		panic(err)
	}

	rt, err := workerruntime.NewApp(cfg)
	if err != nil {
		panic(err)
	}

	dialer := gomail.NewDialer(
		cfg.SenderConfig.Addr,
		cfg.SenderConfig.Port,
		cfg.SenderConfig.From,
		cfg.SenderConfig.Password,
	)

	// infra
	kfkPublisher := kfk.NewPublisher(rt.AsyncWriter)
	kfkPingConsumer := kfk.NewConsumer(rt.PingReader)
	kfkMaillConsumer := kfk.NewConsumer(rt.MailReader)
	gomailSender, err := smtp.NewGomailSender(dialer, rt.Config.SenderConfig.From)
	if err != nil {
		panic(err)
	}

	// service
	downloadService := service.NewDownLoadService(rt.Config.AppConfig.ReportURL, rt.Config.AppConfig.ReportKey, http.DefaultClient)

	var wg sync.WaitGroup
	wg.Add(2)
	go CheckServer(ctx, &wg, rt, kfkPingConsumer, kfkPublisher)
	go SendMail(ctx, &wg, rt, kfkMaillConsumer, gomailSender, downloadService)
	wg.Wait()
}
