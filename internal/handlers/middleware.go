package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ZapLogger(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		lat := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.Duration("latency", lat),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Gin custom_errors (если были)
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("custom_errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			l.Error("http_request", fields...)
		case status >= 400:
			l.Warn("http_request", fields...)
		default:
			l.Info("http_request", fields...)
		}
	}
}

func Recovery(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				l.Error("panic_recovered", zap.Any("panic", r))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			}
		}()
		c.Next()
	}
}

// Простой CORS для локального фронта.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAll := false
	set := map[string]struct{}{}
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o == "*" {
			allowAll = true
			continue
		}
		if o != "" {
			set[o] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowAll && origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" {
			if _, ok := set[origin]; ok {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
