package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitbucket.org/fastbanking/ring-jobscheduler-service/api"
	"bitbucket.org/fastbanking/ring-jobscheduler-service/logger"
	"github.com/gookit/ini/v2/dotenv"
	"github.com/spf13/viper"
)

func main() {
	_loadenv()
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	logger.Configure(logger.LoggerOptions{
		ServiceName: "ring-job-scheduler-service",
		Env:         "local",
		Level:       "DEBUG",
	})
	server := &http.Server{
		Addr:    ":" + viper.GetString("HTTP_PORT"),
		Handler: api.Routes(),
	}

	go func() {
		<-ctx.Done()

		deadlineContext, stop := context.WithTimeout(context.TODO(), 5*time.Second)
		defer stop()
		if err := server.Shutdown(deadlineContext); err != nil {
			logger.Error(deadlineContext, err, "", nil)
		}
	}()
	logger.Info(ctx, "bootstrapping application server", nil)
	server.ListenAndServe()
}

func _loadenv() {
	err := dotenv.Load("./", ".env")
	if err != nil && os.Getenv("SERVICE_NAME") == "" {
		panic(err)
	}
	viper.AutomaticEnv()
}
