package webhook

import (
	"webhook/common"
	"webhook/event"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func LogHandler(c echo.Context) error {
	var payload common.Payload
	if err := c.Bind(&payload); err != nil {
		logrus.WithError(err).Error("Failed to bind payload")
		return c.String(400, "Invalid payload")
	}

	event.WebhookEvent.Add(payload)

	return c.String(200, "Webhook received")
}
