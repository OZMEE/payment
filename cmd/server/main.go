package main

import (
	"context"
	"payment/internal/handler"
	"payment/internal/repository"
	"payment/internal/service"
	"payment/pkg/config"
	"payment/pkg/db"
	"payment/pkg/kafka"
	"payment/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

func main() {
	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	log.Info("Logger initialized")

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log.Info("Configuration loaded",
		zap.Any("config", cfg))

	database, err := db.New(cfg.Database)
	if err != nil {
		panic(err)
	}
	defer func(database *db.Database) {
		err := database.Close()
		if err != nil {
			panic(err)
		}
	}(database)
	log.Info("Successfully connected to db")

	producer, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		panic(err)
	}
	log.Info("Successfully connected to kafka")
	paymentProducerService := service.NewPaymentProducerImpl(cfg.Kafka, producer, log)
	paymentRepo := repository.NewPaymentRepositoryImpl(database)
	outboxRepo := repository.NewOutboxRepositoryImpl(database)
	paymentSvc := service.NewPaymentServiceImpl(paymentRepo, outboxRepo, log)
	workerPool := service.NewWorkerPoolImpl(cfg.Worker, paymentProducerService, outboxRepo, log)
	paymentHandler := handler.NewPaymentHandlerImpl(paymentSvc, log)
	paymentRouter := handler.NewPaymentRouterImpl(paymentHandler)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()

	workerPool.Start(ctx, &wg)

	paymentRouter.Route(cfg.Server)
}
