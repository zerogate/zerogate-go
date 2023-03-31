package zerogate

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tenant ZeroGate tenant
type Tenant struct {
	Base
	AuditBase
	Name         string `json:"name"`
	Description  string `json:"description"`
	Organization string `json:"organization"`
}

// TenantService tenant service
type TenantService service

// TenantCreateRequest tenant create request
type TenantCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TenantUpdateRequest tenant update request
type TenantUpdateRequest struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Create creates a new tenant
func (t *TenantService) Create(ctx context.Context, request *TenantCreateRequest) (*Tenant, error) {
	res, err := t.client.post(ctx, "/tenants", nil, request, nil)
	if err != nil {
		return nil, err
	}
	var r SuccessResponse[*Tenant]
	err = json.Unmarshal(res.Body, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant JSON data: %w", err)
	}
	return r.Data, nil
}

// List get all tenants
func (t *TenantService) List(ctx context.Context) ([]*Tenant, int64, error) {
	res, err := t.client.get(ctx, "/tenants", nil, nil)
	if err != nil {
		return nil, 0, err
	}
	var r SuccessPagingResponse[*Tenant]
	err = json.Unmarshal(res.Body, &r)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal tenant JSON data: %w", err)
	}
	return r.Data, r.Total, nil
}

// Update updates the tenant
func (t *TenantService) Update(ctx context.Context, tenantId string, request *TenantUpdateRequest) (*Tenant, error) {
	res, err := t.client.put(ctx, "/tenants/"+tenantId, nil, request, nil)
	if err != nil {
		return nil, err
	}
	var r SuccessResponse[*Tenant]
	err = json.Unmarshal(res.Body, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant JSON data: %w", err)
	}
	return r.Data, nil
}
