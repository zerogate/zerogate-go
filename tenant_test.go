package zerogate

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestTenantService_Create(t *testing.T) {
	setup()
	defer teardown()
	router.POST("/tenants", func(c *gin.Context) {
		assert.Equal(t, http.MethodPost, c.Request.Method, "Expected method 'POST', got %s", c.Request.Method)
		assert.Equal(t, "application/json", c.Request.Header.Get("Content-Type"))
		testSignature(c, t)
		var json TenantCreateRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			assert.NoError(t, err)
			return
		}
		res := &Tenant{
			Base:         Base{Id: "ten_ea87af463d9fc38203690805c1c1fa33"},
			Name:         json.Name,
			Description:  json.Description,
			Organization: "org_7af4b215d3a00a5dc1f5abf3c3f9686c",
		}
		c.JSON(http.StatusOK, newSuccessResponse(res))
	})
	req := &TenantCreateRequest{
		Name:        "Test",
		Description: "test tenant",
	}
	tenant, err := client.Tenant.Create(context.TODO(), req)
	if err != nil {
		assert.NoError(t, err, "tenant creation error")
		return
	}
	assert.Equal(t, req.Name, tenant.Name, "tenant name is not equal")
	assert.Equal(t, req.Description, tenant.Description, "tenant description is not equal")
	assert.NotEmpty(t, tenant.Id, "tenant id is empty")
	assert.NotEmpty(t, tenant.Organization, "tenant organization is empty")
}

func TestTenantService_List(t *testing.T) {
	setup()
	defer teardown()
	router.GET("/tenants", func(c *gin.Context) {
		assert.Equal(t, http.MethodGet, c.Request.Method, "Expected method 'GET', got %s", c.Request.Method)
		assert.Equal(t, "application/json", c.Request.Header.Get("Content-Type"))
		testSignature(c, t)
		res := []*Tenant{{
			Base:         Base{Id: "ten_ea87af463d9fc38203690805c1c1fa33"},
			Name:         "Test",
			Description:  "test tenant",
			Organization: "org_7af4b215d3a00a5dc1f5abf3c3f9686c",
		}}
		c.JSON(http.StatusOK, newSuccessPagingResponse(res, 1))
	})

	tenants, total, err := client.Tenant.List(context.TODO())
	if err != nil {
		assert.NoError(t, err, "tenant creation error")
		return
	}
	assert.Equal(t, len(tenants), 1, "tenants length should be 1")
	assert.Equal(t, total, int64(1), "tenants length should be 1")
	tenant := tenants[0]
	assert.Equal(t, "Test", tenant.Name, "tenant name is not equal")
	assert.Equal(t, "test tenant", tenant.Description, "tenant description is not equal")
	assert.NotEmpty(t, tenant.Id, "tenant id is empty")
	assert.NotEmpty(t, tenant.Organization, "tenant organization is empty")
}

func TestTenantService_Update(t *testing.T) {
	setup()
	defer teardown()
	router.PUT("/tenants/:tenantId", func(c *gin.Context) {
		assert.Equal(t, http.MethodPut, c.Request.Method, "Expected method 'PUT', got %s", c.Request.Method)
		assert.Equal(t, "application/json", c.Request.Header.Get("Content-Type"))
		testSignature(c, t)
		var json TenantUpdateRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			assert.NoError(t, err)
			return
		}
		res := &Tenant{
			Base:         Base{Id: json.Id},
			Name:         json.Name,
			Description:  json.Description,
			Organization: "org_7af4b215d3a00a5dc1f5abf3c3f9686c",
		}
		c.JSON(http.StatusOK, newSuccessResponse(res))
	})
	req := &TenantUpdateRequest{
		Id:          "ten_ea87af463d9fc38203690805c1c1fa33",
		Name:        "Test Update",
		Description: "test tenant update",
	}
	tenant, err := client.Tenant.Update(context.TODO(), req.Id, req)
	if err != nil {
		assert.NoError(t, err, "tenant update error")
		return
	}
	assert.Equal(t, req.Name, tenant.Name, "tenant name is not equal")
	assert.Equal(t, req.Description, tenant.Description, "tenant description is not equal")
	assert.Equal(t, req.Id, tenant.Id, "tenant id is not equal")
	assert.NotEmpty(t, tenant.Organization, "tenant organization is empty")
}
