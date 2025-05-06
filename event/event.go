package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
	"webhook/common"
	"webhook/misc/config"

	"github.com/sirupsen/logrus"
)

type Event struct {
	payloads     []common.Payload
	mutex        sync.Mutex
	lastSentTime time.Time
	EventCtx     context.Context
	EventCancel  context.CancelFunc
}

var (
	WebhookEvent = newEvent()
)

func newEvent() *Event {
	ctx, cancel := context.WithCancel(context.Background())
	return &Event{
		payloads:     make([]common.Payload, 0),
		lastSentTime: time.Now(),
		EventCtx:     ctx,
		EventCancel:  cancel,
	}
}

func (b *Event) Add(payload common.Payload) {
	env := config.GetEnv()
	if env.App == nil {
		logrus.Error("config not initialized")
		return
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.payloads = append(b.payloads, payload)

	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload added to batch")

	if len(b.payloads) >= env.App.BatchSize {
		go b.ForwardPayloads()
	}
}

func (b *Event) ForwardPayloads() {
	fmt.Println("Forwarding payloads")
	b.mutex.Lock()
	if len(b.payloads) == 0 {
		fmt.Println("No payloads to send")
		b.mutex.Unlock()
		return
	}

	payloadsToSend := make([]common.Payload, len(b.payloads))
	copy(payloadsToSend, b.payloads)
	fmt.Println("Payloads to send:", payloadsToSend)

	b.payloads = make([]common.Payload, 0)
	b.lastSentTime = time.Now()
	b.mutex.Unlock()

	jsonData, err := json.Marshal(payloadsToSend)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal payloads")
		return
	}

	fmt.Println("Sending payloads:", string(jsonData))
	b.sendWithRetry(jsonData, len(payloadsToSend))
}

func (b *Event) sendWithRetry(jsonData []byte, batchSize int) {
	maxRetries := 3
	retryCount := 0
	var statusCode int
	var err error
	var duration time.Duration

	for retryCount < maxRetries {
		startTime := time.Now()
		statusCode, err = b.sendPayloads(jsonData)
		duration = time.Since(startTime)
		fmt.Println("Status code:", statusCode)
		fmt.Println("Duration:", duration)

		logrus.WithFields(logrus.Fields{
			"attempt":     retryCount + 1,
			"batch_size":  batchSize,
			"status_code": statusCode,
			"duration":    duration,
		}).Info("Batch forwarding attempt")

		if err == nil && statusCode >= 200 && statusCode < 300 {
			logrus.WithFields(logrus.Fields{
				"status_code": statusCode,
				"duration":    duration,
			}).Info("Batch forwarded successfully")

			return
		}

		retryCount++
		if retryCount < maxRetries {
			logrus.WithFields(logrus.Fields{
				"retry":        retryCount,
				"max_retries":  maxRetries,
				"delay":        2 * time.Second,
				"current_time": time.Now(),
			}).Warn("Retrying batch forwarding")

			time.Sleep(2 * time.Second)
		}
	}

	logrus.WithFields(logrus.Fields{
		"max_retries":  maxRetries,
		"current_time": time.Now(),
	}).Error("Failed to forward batch after retries")

	os.Exit(1)
}

func (b *Event) sendPayloads(jsonData []byte) (int, error) {
	fmt.Println("Sending payloads to endpoint")
	env := config.GetEnv()
	if env.App == nil {
		return 0, fmt.Errorf("config not initialized")
	}
	endpoint := env.App.Endpoint
	if !(len(endpoint) >= 7 && (endpoint[:7] == "http://" || (len(endpoint) >= 8 && endpoint[:8] == "https://"))) {
		logrus.WithField("endpoint", endpoint).Warn("Endpoint missing protocol, prepending http://")
		endpoint = "http://" + endpoint
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	fmt.Println("Response status code:", resp.StatusCode)
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (b *Event) StartIntervalTimer(ctx context.Context) {
	env := config.GetEnv()
	if env.App == nil {
		logrus.Error("config not initialized")
		return
	}
	ticker := time.NewTicker(env.App.BatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			timeSinceLastSend := time.Since(b.lastSentTime)
			if timeSinceLastSend >= env.App.BatchInterval {
				logrus.WithFields(logrus.Fields{
					"interval":             env.App.BatchInterval,
					"time_since_last_send": timeSinceLastSend,
				}).Info("Batch interval triggered")
				b.ForwardPayloads()
			}
		case <-ctx.Done():
			return
		}
	}
}
