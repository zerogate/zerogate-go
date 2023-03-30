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

// Response response object
type Response[T any] struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Success      bool   `json:"success"`
	Data         T      `json:"data"`
	Total        int64  `json:"total"`
}

// APIResponse API response
type APIResponse struct {
	Body       []byte
	Status     string
	StatusCode int
	Headers    http.Header
}
