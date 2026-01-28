package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"github.com/tomiwa-a/Relay/internal/api/app"
	"github.com/tomiwa-a/Relay/internal/api/routes"
	"github.com/tomiwa-a/Relay/internal/worker"
)

func main() {

	config := app.LoadConfig()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := app.OpenDB(config.DB)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(config.Kafka.Brokers...),
		Topic:    config.Kafka.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	defer kafkaWriter.Close()

	application := app.NewApplication(config, logger, db, kafkaWriter)

	backgroundWorker := worker.NewWorker(application)

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	go backgroundWorker.Start(workerCtx)

	r := gin.Default()

	routes.HandleRequests(r, application)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Printf("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	logger.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Printf("server forced to shutdown: %v", err)
	}

	logger.Println("server stopped")
}
