package custom_errors

import "errors"

var (
	ErrFailedToOpenFile = errors.New("failed to open file")
	ErrFailedToReadFile = errors.New("failed to read file")
	ErrInvalidXLSX      = errors.New("invalid xlsx")
	ErrNoXLSXSheets     = errors.New("no xlsx sheets found")
	ErrNoXLSXData       = errors.New("no xlsx data")
	ErrTimeFormat       = errors.New("error time format")
)
