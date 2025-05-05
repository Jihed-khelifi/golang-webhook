package webhook

import "github.com/labstack/echo/v4"

func RegisterRouter(server *echo.Group) {

	server.GET(("/healthz"), func(c echo.Context) error {
		return c.String(200, "OK")
	})

	server.POST("/log", LogHandler)
}
