package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kurushqosimi/x5-intern-hiring/internal/custom_errors"
	"go.uber.org/zap"
	"net/http"
)

// curl -F "file=@пример выгрузки отклика.xlsx" http://localhost:8080/api/v1/imports/xlsx

func (h *Handler) UploadXLSX(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		h.logger.Error("file's absent", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "файл не был загружен"})
		return
	}

	procResult, err := h.service.ProcessXLSX(ctx, fileHeader)
	if err != nil {
		h.logger.Error("h.service.ProcessXLSX: ", zap.Error(err))
		switch {
		case errors.Is(err, custom_errors.ErrFailedToOpenFile):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "невозможно открыть загруженнный xlsx"})
		case errors.Is(err, custom_errors.ErrFailedToReadFile):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "невозможно прочесть файл"})
		case errors.Is(err, custom_errors.ErrInvalidXLSX):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "невалидный xlsx"})
		case errors.Is(err, custom_errors.ErrNoXLSXSheets):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "xlsx не имеет страниц"})
		case errors.Is(err, custom_errors.ErrNoXLSXData):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "xlsx не имеет данных"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"import_id":     procResult.ImportId,
		"file_sha256":   procResult.FileSha256,
		"total_rows":    procResult.TotalRows,
		"inserted_rows": procResult.InsertedRows,
		"skipped_rows":  procResult.SkippedRows,
		"errors":        procResult.Errors,
	})
}
