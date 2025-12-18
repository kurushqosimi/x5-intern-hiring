package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kurushqosimi/x5-intern-hiring/internal/services"
	"github.com/kurushqosimi/x5-intern-hiring/pkg/logger"
	"go.uber.org/zap"
)

var localAddresses = []string{
	"http://localhost:5173",
	"http://127.0.0.1:5173",
	"http://localhost:3000",
	"http://127.0.0.1:3000",
}

type Handler struct {
	logger  *zap.Logger
	service *services.Service
}

func NewHandler(logger *zap.Logger, service *services.Service) *Handler {
	return &Handler{logger: logger, service: service}
}

const (
	importsXLSX = "/imports/xlsx"
)

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.New()
	r.Use(Recovery(logger.Get()))
	r.Use(ZapLogger(logger.Get()))
	r.Use(CORS(localAddresses))

	api := r.Group("/api/v1")
	api.POST(importsXLSX, h.UploadXLSX)

	return r
}
