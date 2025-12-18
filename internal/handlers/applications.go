package handlers

import (
	"github.com/kurushqosimi/x5-intern-hiring/internal/custom_errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kurushqosimi/x5-intern-hiring/internal/models"
)

// curl "http://localhost:8080/api/v1/applications?limit=20&offset=0&status=NEW,IN_REVIEW&q=Петр"
func (h *Handler) ListApplications(ctx *gin.Context) {
	p := models.ListApplicationsParams{
		Limit:  parseInt(ctx.Query("limit"), 50),
		Offset: parseInt(ctx.Query("offset"), 0),

		Q:          ctx.Query("q"),
		Priority:   ctx.Query("priority"),
		Course:     ctx.Query("course"),
		Specialty:  ctx.Query("specialty"),
		Schedule:   ctx.Query("schedule"),
		City:       ctx.Query("city"),
		University: ctx.Query("university"),

		Citizenship: ctx.Query("citizenship"),
		ImportID:    ctx.Query("import_id"),
	}

	// status=NEW,INVITED,CRM_QUEUED...
	if s := strings.TrimSpace(ctx.Query("status")); s != "" {
		parts := strings.Split(s, ",")
		for _, x := range parts {
			x = strings.TrimSpace(x)
			if x != "" {
				p.Statuses = append(p.Statuses, x)
			}
		}
	}

	// applied_from / applied_to: RFC3339 или YYYY-MM-DD
	if v := strings.TrimSpace(ctx.Query("applied_from")); v != "" {
		if t, err := parseTime(v); err == nil {
			p.AppliedFrom = &t
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid applied_from"})
			return
		}
	}
	if v := strings.TrimSpace(ctx.Query("applied_to")); v != "" {
		if t, err := parseTime(v); err == nil {
			p.AppliedTo = &t
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid applied_to"})
			return
		}
	}

	// has_resume=true/false
	if v := strings.TrimSpace(ctx.Query("has_resume")); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid has_resume"})
			return
		}
		p.HasResume = &b
	}

	res, err := h.service.ListApplications(ctx.Request.Context(), p)
	if err != nil {
		h.logger.Error("h.service.ListApplications: ", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func parseInt(s string, def int) int {
	if strings.TrimSpace(s) == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func parseTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, custom_errors.ErrTimeFormat
}
