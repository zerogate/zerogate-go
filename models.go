package zerogate

import (
	"database/sql"
	"net/http"
)

// Base common model
type Base struct {
	Id        string       `json:"id"`
	Updated   int64        `json:"created"`
	Created   int64        `json:"updated"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

// TenantBase tenant model
type TenantBase struct {
	Base
	Tenant string `json:"tenant"`
}

// AuditBase audit model
type AuditBase struct {
	CreatedBy  string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
}

// SuccessResponse success response
type SuccessResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

// SuccessPagingResponse success paging response
type SuccessPagingResponse[T any] struct {
	Success bool  `json:"success"`
	Data    []T   `json:"data"`
	Total   int64 `json:"total"`
}

// ErrorResponse error response
type ErrorResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Success      bool   `json:"success"`
}

// APIResponse API response
type APIResponse struct {
	Body       []byte
	Status     string
	StatusCode int
	Headers    http.Header
}

func newSuccessResponse[T any](data T) *SuccessResponse[T] {
	return &SuccessResponse[T]{
		Success: true,
		Data:    data,
	}
}

func newSuccessPagingResponse[T any](data []T, total int64) *SuccessPagingResponse[T] {
	return &SuccessPagingResponse[T]{
		Success: true,
		Data:    data,
		Total:   total,
	}
}

func newErrorResponse(code int, message error) *ErrorResponse {
	return &ErrorResponse{
		Success:      false,
		ErrorCode:    code,
		ErrorMessage: message.Error(),
	}
}
func newErrorsResponse(code int, messages string) *ErrorResponse {
	return &ErrorResponse{
		Success:      false,
		ErrorCode:    code,
		ErrorMessage: messages,
	}
}
