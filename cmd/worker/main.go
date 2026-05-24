package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	workerconfig "github.com/LeHuuHai/server-management/config/worker"
	"github.com/LeHuuHai/server-management/internal/domain/mail"
	"github.com/LeHuuHai/server-management/internal/domain/mq"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	smtp "github.com/LeHuuHai/server-management/internal/infra/mail"
	workerruntime "github.com/LeHuuHai/server-management/internal/infra/runtime/worker"
	"github.com/LeHuuHai/server-management/internal/model"
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
			if msg.Topic != rt.Config.KafkaReaderConfig.PingTopic {
				log.Printf("Received message with topic %s, expected %s", msg.Topic, rt.Config.KafkaReaderConfig.PingTopic)
				continue
			}
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
						IP:     req.IP,
						Status: "on",
						PingAt: time.Now(),
					}
					conn, err := net.DialTimeout(
						"tcp",
						net.JoinHostPort(req.IP, "22"),
						1*time.Second,
					)
					if err != nil {
						res.Status = "off"
					}
					resBytes, err := json.Marshal(res)
					if err != nil {
						log.Println(err.Error())
						continue
					}
					msg := mq.Message{
						Topic: "ping_res",
						Value: resBytes,
					}
					err = publisher.Publish(ctx, msg)
					if err != nil {
						log.Println(err.Error())
					}
					if conn != nil {
						_ = conn.Close()
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
) {
	defer wg.Done()
	for {
		msg, err := consumer.Read(ctx)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if msg.Topic != rt.Config.KafkaReaderConfig.MailTopic {
			log.Printf("Received message with topic %s, expected %s", msg.Topic, rt.Config.KafkaReaderConfig.MailTopic)
			continue
		}
		var mailReq model.RequestMail
		err = json.Unmarshal(msg.Value, &mailReq)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		err = sender.Send(ctx, mailReq.Mail)
		if err != nil {
			log.Println(err.Error())
			continue
		}
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
	kfkPublisher := kfk.NewPublisher(rt.SyncWriter)
	kfkPingConsumer := kfk.NewConsumer(rt.PingReader)
	kfkMaillConsumer := kfk.NewConsumer(rt.MailReader)
	gomailSender, err := smtp.NewGomailSender(dialer, rt.Config.SenderConfig.From)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go CheckServer(ctx, &wg, rt, kfkPingConsumer, kfkPublisher)
	go SendMail(ctx, &wg, rt, kfkMaillConsumer, gomailSender)
	wg.Wait()
}
