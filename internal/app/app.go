package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kurushqosimi/x5-intern-hiring/internal/handlers"
	"github.com/kurushqosimi/x5-intern-hiring/internal/repositories"
	"github.com/kurushqosimi/x5-intern-hiring/internal/services"
	"github.com/kurushqosimi/x5-intern-hiring/pkg/db/postgres"
	"github.com/kurushqosimi/x5-intern-hiring/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	pgDSN    = "postgres://intern_hiring_user:intern_hiring_password@postgres:5432/intern_hiring_db?sslmode=disable"
	httpAddr = ":8080"
)

func Start() {
	l := logger.New()
	defer func() {
		_ = l.Sync()
	}()

	ctx := context.Background()
	pool, err := postgres.New(ctx, pgDSN)
	if err != nil {
		l.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	repo := repositories.NewRepository(pool)
	service := services.NewService(repo)
	handler := handlers.NewHandler(l, service)

	router := handler.InitRoutes()
	srv := &httpServer{engine: router, addr: httpAddr, l: l}
	go func() {
		if err := srv.Run(); err != nil {
			l.Fatal("server failed", zap.Error(err))
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shctx)

}

type httpServer struct {
	engine *gin.Engine
	addr   string
	l      *zap.Logger
	s      *http.Server
}

func (h *httpServer) Run() error {
	h.s = &http.Server{
		Addr:    h.addr,
		Handler: h.engine,
	}
	h.l.Info("http_listen", zap.String("addr", h.addr))
	return h.s.ListenAndServe()
}

func (h *httpServer) Shutdown(ctx context.Context) error {
	h.l.Info("http_shutdown")
	return h.s.Shutdown(ctx)
}
