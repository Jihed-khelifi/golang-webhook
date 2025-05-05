package server

import (
	"webhook/misc/config"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

func logRecoverPanics(c echo.Context, err error, stack []byte) error {
	stackString := ""

	if config.IsDevelopment() {
		stackString = string(stack)
	}

	logrus.WithFields(logrus.Fields{
		"error": err,
		"stack": stackString,
	}).Error("panic recovered")

	return c.String(500, "Internal Server Error")
}

func BootStrap() *echo.Echo {
	log := logrus.New()

	e := echo.New()

	e.Debug = config.IsDevelopment()

	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.WithFields(logrus.Fields{
					"URI":    v.URI,
					"status": v.Status,
				}).Info("request")
			} else {
				log.WithFields(logrus.Fields{
					"URI":    v.URI,
					"status": v.Status,
					"error":  v.Error,
				}).Error("request error")
			}
			return nil
		},
	}))

	recoverConfig := echomiddleware.RecoverConfig{
		Skipper:             echomiddleware.DefaultSkipper,
		StackSize:           4 << 10, // 4 KB
		DisableStackAll:     false,
		DisablePrintStack:   false,
		LogLevel:            0,
		LogErrorFunc:        logRecoverPanics,
		DisableErrorHandler: false,
	}
	e.Use(echomiddleware.RecoverWithConfig(recoverConfig))

	return e
}
