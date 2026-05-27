package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"runtime"
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
				log.Println(err.Error())
				continue
			}
			consumer.Commit(ctx, msg)
			var pingReq model.RequestPing
			err = json.Unmarshal(msg.Value, &pingReq)
			if err != nil {
				log.Println(err.Error())
				continue
			}
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
						Status:   "off",
						PingAt:   time.Now(),
					}
					var cmd *exec.Cmd
					if runtime.GOOS == "windows" {
						// Windows: -n 1 (gửi 1 gói tin), -w 1000 (thời gian chờ 1000ms)
						cmd = exec.Command("ping", "-n", "1", "-w", "1000", req.IP)
					} else {
						// Linux/macOS: -c 1 (gửi 1 gói tin), -W 1 (thời gian chờ 1 giây)
						cmd = exec.Command("ping", "-c", "1", "-W", "1", req.IP)
					}
					// Chạy lệnh và kiểm tra kết quả trả về
					err := cmd.Run()
					// Nếu lệnh ping chạy thành công (trả về exit code 0) tức là server có phản hồi -> ON
					if err == nil {
						res.Status = "on"
					}
					resBytes, err := json.Marshal(res)
					if err != nil {
						log.Println(err.Error())
						continue
					}
					msg := mq.Message{
						Topic: rt.Config.KafkaConfig.Topics["ping_res"],
						Value: resBytes,
					}
					err = publisher.Publish(ctx, msg)
					if err != nil {
						log.Println(err.Error())
					}
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
			log.Println(err.Error())
			continue
		}
		var mailReq model.RequestMail
		err = json.Unmarshal(msg.Value, &mailReq)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		// attachments
		valid := true
		for i, attachment := range mailReq.Mail.Attachments {
			data, err := downloadService.Download(
				ctx,
				attachment.Filename,
			)
			if err != nil {
				log.Printf(
					"download attachment %s failed: %v",
					attachment.Filename,
					err,
				)
				valid = false
				break
			}

			mailReq.Mail.Attachments[i].Data = data
		}
		if !valid {
			log.Println("cannot send mail because of miss attachment")
			continue
		}
		err = sender.Send(ctx, mailReq.Mail)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		consumer.Commit(ctx, msg)
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
	downloadService := service.NewDownLoadService(rt.Config.AppConfig.ReportURL, http.DefaultClient)

	var wg sync.WaitGroup
	wg.Add(2)
	go CheckServer(ctx, &wg, rt, kfkPingConsumer, kfkPublisher)
	go SendMail(ctx, &wg, rt, kfkMaillConsumer, gomailSender, downloadService)
	wg.Wait()
}
