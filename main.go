package main

import (
	"context"
	"fmt"
	"go_web_scaffold/dao/mysql"
	"go_web_scaffold/dao/redis"
	"go_web_scaffold/logger"
	"go_web_scaffold/routes"
	"go_web_scaffold/settings"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 1. load config
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err: %v\n", err)
		return
	}

	// 2. init log
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed, err: %v\n", err)
		return
	}
	// Refresh all buffered logs synchronously
	defer func() {
		_ = zap.L().Sync()
	}()

	// 3. init mysql
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err: %v\n", err)
		return
	}

	// close mysql
	defer mysql.Close()

	// 4. init redis
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err: %v\n", err)
		return
	}

	// close redis
	defer redis.Close()

	// 5. setup routes
	r := routes.Setup()

	// 6. start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: r,
	}

	go func() {
		// Start a goroutine to start the serve
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for the interrupt signal to gracefully shut down the server,
	// and set a 5 second timeout for the shutdown operation
	quit := make(chan os.Signal, 1) // Create a channel to receive the signal
	// kill The syscall.SIGTERM signal is sent by default
	// kill -2 send syscall.SIGINT signal，our commonly used Ctrl+C is to trigger the system SIGINT signal
	// kill -9 send syscall.SIGKILL signal，But it cannot be captured, so there is no need to add it
	// signal.Notify forwards the received syscall.SIGINT or syscall.SIGTERM signal to quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Won't block here
	<-quit                                               // Blocking here, only when the above two signals are received will execute
	zap.L().Info("Shutdown Server ...")
	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Gracefully shut down the service within 5 seconds (close the service after processing unprocessed requests),
	// and exit after 5 seconds.
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
