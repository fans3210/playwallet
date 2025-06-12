package apis

import (
	"context"
	"log/slog"
	"net"

	"playwallet/internal/biz"
	"playwallet/internal/cfgs"
	"playwallet/internal/data"
	"playwallet/pkg/middlewares"

	"github.com/labstack/echo/v4"
)

type App struct {
	cfg cfgs.Config
	svr *echo.Echo
	uc  *biz.WalletUC
}

func NewApp(cfg cfgs.Config) (*App, error) {
	s := &App{
		cfg: cfg,
	}
	// db
	repo, err := data.NewWalletRepo(cfg.PG)
	if err != nil {
		return nil, err
	}
	slog.Info("connect to postgres successfully with db", "db", cfg.PG.DB)
	// biz
	uc, err := biz.NewWalletUC(cfg, repo)
	if err != nil {
		return nil, err
	}
	s.uc = uc

	// routes & middleware
	s.svr = echo.New()
	s.svr.HideBanner = cfg.Env == "test"
	// middlewares & routes
	s.svr.Use(middlewares.ErrorConvMiddleware)
	s.registerRoutes()

	return s, nil
}

// specify :0 to start with a random port, for easier testing
func (s *App) Start() error {
	ln, err := s.NewListener()
	if err != nil {
		return err
	}
	return s.StartWithListener(ln)
}

func (s *App) NewListener() (net.Listener, error) {
	ln, err := net.Listen("tcp", s.cfg.Http.Addr)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

func (s *App) StartWithListener(ln net.Listener) error {
	slog.Debug("server starting...")
	s.svr.Listener = ln
	return s.svr.Start(s.cfg.Http.Addr)
}

func (s *App) ShunDown() error {
	return s.svr.Shutdown(context.Background())
}

func (s *App) registerRoutes() {
	slog.Debug("registering routes...")

	s.svr.GET("/balance/:userid", s.getBalacne)
	s.svr.POST("/deposit", s.deposit)
}
