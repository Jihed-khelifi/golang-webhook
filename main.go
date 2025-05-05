package main

import (
	"fmt"
	"os"
	"webhook/event"
	"webhook/misc/config"
	"webhook/server"
	"webhook/webhook"

	"github.com/sirupsen/logrus"
)

type Rest struct{}

func main() {
	env := config.GetEnv()
	if err := env.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup environment: %v\n", err)
		os.Exit(1)
	}

	server := server.BootStrap()

	group := server.Group("/v1")

	webhook.RegisterRouter(group)

	// Start the batch interval timer
	go event.WebhookEvent.StartIntervalTimer(event.WebhookEvent.EventCtx)

	logrus.WithFields(logrus.Fields{
		"port":           env.App.Port,
		"endpoint":       env.App.Endpoint,
		"batch_size":     env.App.BatchSize,
		"batch_interval": env.App.BatchInterval,
	}).Info("Application started")

	if err := server.Start(fmt.Sprintf(":%d", env.App.Port)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

}
