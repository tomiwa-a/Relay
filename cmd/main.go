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
	"github.com/tomiwa-a/Relay/internal/api/app"
	"github.com/tomiwa-a/Relay/internal/api/routes"
)

func main() {

	a := 1
	str := fmt.Sprintf("the value of a is %d", a)
	fmt.Println(str)

	arr := [4]int{1, 2, 3}

	fmt.Println(arr)

	return

	config := app.LoadConfig()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := app.OpenDB(config.DB)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	app := app.NewApplication(config, logger, db)

	r := gin.Default()

	routes.HandleRequests(r, app)

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
