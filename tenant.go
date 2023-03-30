package zerogate

type Tenant struct {
	Base
	AuditBase
	Name         string `json:"name"`
	Description  string `json:"description"`
	Organization string `json:"organization"`
}
