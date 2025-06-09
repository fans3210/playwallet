package apis

import (
	"log"
	"log/slog"

	"playwallet/pkg/middlewares"

	"github.com/labstack/echo/v4"
)

type Server struct {
	svr *echo.Echo
}

func NewServer() (*Server, error) {
	s := &Server{}
	s.svr = echo.New()
	// middlewares
	s.svr.Use(middlewares.ErrorConvMiddleware)

	s.registerRoutes()
	return s, nil
}

func (s *Server) Start() {
	slog.Info("hi there starting a new server")
	slog.Debug("debug hi there starting a new server")
	log.Fatal(s.svr.Start(":1323"))
}

func (s *Server) registerRoutes() {
	s.svr.GET("/hello", func(c echo.Context) error {
		return nil
	})
}
